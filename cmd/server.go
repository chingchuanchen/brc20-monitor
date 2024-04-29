package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github/HyprNetwork/brc20-balance-monitor/constant"
	"github/HyprNetwork/brc20-balance-monitor/model"
	"github/HyprNetwork/brc20-balance-monitor/service"
	"github/HyprNetwork/brc20-balance-monitor/utils"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	var network = flag.String("network", "testnet", "network type")
	flag.Parse()

	done := make(chan struct{})
	defer close(done)

	interval, err := strconv.Atoi(os.Getenv(constant.Interval))
	if err != nil {
		utils.GetScanLogger().Fatalf("get interval err: %v", err)
		return
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	startBlockHeight, err := strconv.Atoi(os.Getenv(constant.StartBlock))
	if err != nil {
		utils.GetScanLogger().Fatalf("get startblock height err: %v", err)
		return
	}

	latestBlock, err := model.GetLatestBlock()
	if err != nil {
		utils.GetScanLogger().Fatalf("Handle GetLatestBlock err: %v", err)
		return
	}

	if latestBlock.Height != 0 {
		// 不是第一次scan
		startBlockHeight = int(latestBlock.Height)
	}

	for {
		select {
		case <-done:
			utils.GetScanLogger().Info("scan service done")
			return
		case <-ticker.C:
			utils.GetScanLogger().Infof("Start scan block %d...", startBlockHeight)
			startBlock, err := model.NewBlockFromDB(int64(startBlockHeight))
			if err != nil {
				utils.GetScanLogger().Fatalf("get startblock err: %v", err)
				break
			}
			if startBlock.IsSyncFinished() && latestBlock.Height >= int64(startBlockHeight) {
				utils.GetScanLogger().Errorf("current block height should bigger, lastest is: %d, current is: %d", latestBlock.Height, startBlockHeight)
				startBlockHeight += 1
				break
			}
			url := fmt.Sprintf("https://prod-%s.prod.findora.org:26657/block?height=%d", *network, startBlockHeight)
			utils.GetScanLogger().Infof("request %s", url)
			resp, err := http.Get(url)
			if err != nil {
				utils.GetScanLogger().Errorf("get request err: %+v", err)
				break
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				utils.GetScanLogger().Errorf("Error reading response body: %+v", err)
				break
			}

			var blockResp model.BlockResponse
			err = json.Unmarshal(body, &blockResp)
			if err != nil {
				utils.GetScanLogger().Errorf("unmarsh resp body err: %+v", err)
				break
			}
			if blockResp.Error != nil {
				utils.GetScanLogger().Errorf("request is error, content: %v", blockResp.Error)
				break
			}
			txHandle := service.NewTxHandler(blockResp.Result.Block)
			err = txHandle.Handle(int64(startBlockHeight))
			if err != nil {
				utils.GetScanLogger().Errorf("handle block msg err: %+v", err)
				startBlockHeight += 1
				break
			}
			block := &model.Block{Height: int64(startBlockHeight), IsFinished: true}
			err = block.InsertToDB()
			if err != nil {
				utils.GetScanLogger().Errorf("insert block to db err: %+v", err)
				startBlockHeight += 1
				break
			}
			startBlockHeight += 1
		}
	}
}
