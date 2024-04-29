package service

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github/HyprNetwork/brc20-balance-monitor/decimal"
	"github/HyprNetwork/brc20-balance-monitor/model"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Transaction struct{}

func (t *Transaction) GetResponse(reqUrl string) (*model.TxResponse, error) {

	resp, err := http.Get(reqUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var txResp model.TxResponse
	err = json.Unmarshal(body, &txResp)
	if err != nil {
		return nil, err
	}

	return &txResp, nil
}

func (t *Transaction) CheckBuy(reqUrl string, buyer string, center string, fraTotal string) (bool, error) {
	resp, err := t.GetResponse(reqUrl)
	if err != nil {
		return false, err
	}

	txBytes, err := base64.StdEncoding.DecodeString(resp.Result.Tx)
	if err != nil {
		return false, err
	}
	if strings.Contains(string(txBytes), "evm") {
		return false, errors.New("error tx")
	}
	var operations model.HanldeBody
	err = json.Unmarshal(txBytes, &operations)
	if err != nil {
		return false, err
	}

	fraTotalDecimal, _, err := decimal.NewDecimalFromString(fraTotal)
	if err != nil {
		return false, err
	}

	for _, op := range operations.Body.Operations {
		if len(op.TransferAsset.Body.Outputs) == 0 {
			continue
		}

		for _, output := range op.TransferAsset.Body.Outputs {
			amt, err := strconv.ParseInt(output.Record.Amount.NonConfidential, 10, 64)
			if err != nil {
				return false, err
			}
			if output.Record.PublicKey == center && amt == fraTotalDecimal.Value.Int64() {
				return true, nil
			}
		}
	}

	return false, errors.New("error tx")
}
