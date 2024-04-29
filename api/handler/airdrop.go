package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github/HyprNetwork/brc20-balance-monitor/constant"
	"github/HyprNetwork/brc20-balance-monitor/model"
	"github/HyprNetwork/brc20-balance-monitor/utils"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func AirDropCheckAndSend(ctx *gin.Context) {
	unisatUser := ctx.PostForm("walletUser")
	fraUser := ctx.PostForm("fraUser")
	pubKey := ctx.PostForm("walletPubKey")
	signature := ctx.PostForm("signature")

	// 1. 验证签名
	fraPubKey, err := utils.GetPubkeyFromAddress(fraUser)
	if err != nil {
		utils.GetLogger().Errorf("GetPubkeyFromAddress err:%+v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}
	ok := utils.VerifyMessage(pubKey, fraUser, signature)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.New("invalid signature").Error())
		return
	}

	// 2. 查询ORDI balance
	// https://open-api.unisat.io/v1/indexer/address/{address}/brc20/summary
	client := &http.Client{}

	// 创建一个新的 HTTP 请求
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/indexer/address/%s/brc20/summary", os.Getenv(constant.UNISATDOMAIN), unisatUser), nil)
	if err != nil {
		utils.GetLogger().Errorf("NewRequest err:%v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	// 添加自定义的 Header 头部
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv(constant.UNISATAPIKEY)))
	req.Header.Add("Accept", "application/json")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		utils.GetLogger().Errorf("Request unisat err:%v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	defer resp.Body.Close()

	// 处理响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.GetLogger().Errorf("Request unisat err:%v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	var balanceResp model.UnisatResponse[model.UnisatTickerSummary]
	err = json.Unmarshal(body, &balanceResp)
	if err != nil {
		utils.GetLogger().Errorf("Unmarshal UnisatResponse err:%v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	if len(balanceResp.Data.Detail) == 0 {
		ctx.JSON(http.StatusOK, gin.H{"result": "false"})
		return
	}

	validateMap := map[string]string{
		"ordi": "1",
		"sats": "100000",
		"rats": "1000",
		"tbci": "1000",
		"ainn": "21",
		"ligo": "1000",
		"insc": "1",
		"piin": "1000",
		"slor": "20",
		"zbit": "2100",
	}

	canDrop := false
	for _, v := range balanceResp.Data.Detail {
		if value, ok := validateMap[v.Ticker]; ok {
			if v.OverallBalance >= value {
				canDrop = true
				break
			}
		}
	}

	if canDrop {
		// 空投数额
		amount := os.Getenv(constant.AIRDROPAMOUNT)

		// 发送空投
		_, err = utils.Transfer(os.Getenv(constant.AIRDROPMNEMONIC), fraPubKey, amount)
		if err != nil {
			utils.GetLogger().Errorf("Transfer airdrop err:%v", err)
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
			return
		}

		// 记录空投
		airdrop := &model.AirdropRecord{
			FromUser: os.Getenv(constant.AIRDROPUSER),
			ToUser:   fraUser,
			Amount:   amount,
		}
		err = airdrop.InsertToDB()
		if err != nil {
			utils.GetLogger().Errorf("Insert airdrop err:%v", err)
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"result": "ok", "message": "Congrats! You qualify for an FRA airdrop based on the number of ORDI in your wallet."})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"result": "false"})
	}
}
