package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github/HyprNetwork/brc20-balance-monitor/constant"
	"github/HyprNetwork/brc20-balance-monitor/model"
	"github/HyprNetwork/brc20-balance-monitor/platform"
	"io"
	"net/http"
	"os"
)

func GetOwnedUTXO(pubkey string, endpoint string) (sid uint64, record []byte, err error) {
	resp, err := http.Get(fmt.Sprintf("%s/owned_utxos/%s", endpoint, pubkey))
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	return
}

func SendTx(from string, receiverPubkey string, toPubkey string, amount string, tick string, fraPrice string, brcType string) (string, error) {
	txJsonString := platform.GetTxBody([]byte(from), []byte(receiverPubkey), []byte(toPubkey), []byte(fmt.Sprintf("%s:%s", os.Getenv(constant.ENDPOINT), os.Getenv(constant.PLATINNERPORT))), []byte(amount), []byte(tick), []byte(fraPrice), []byte(brcType))
	result_tx := base64.URLEncoding.EncodeToString([]byte(txJsonString))
	return sendRequest(result_tx)
}

func sendRequest(result_tx string) (string, error) {
	req := model.NewBlockRequest(result_tx)
	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(fmt.Sprintf("%s:%s", os.Getenv(constant.ENDPOINT), os.Getenv(constant.PLATAPIPORT)), "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("request error")
	}

	return string(respBody), nil
}

func Transfer(from string, receiverPubkey string, amount string) (string, error) {
	txJsonString := platform.GetTransferBody([]byte(from), []byte(receiverPubkey), []byte(amount), []byte(fmt.Sprintf("%s:%s", os.Getenv(constant.ENDPOINT), os.Getenv(constant.PLATINNERPORT))))
	result_tx := base64.URLEncoding.EncodeToString([]byte(txJsonString))
	return sendRequest(result_tx)
}

func SendRobotBatch(from string) (string, error) {
	txJsonString := platform.GetSendRobotBatchTxBody([]byte(from), []byte(fmt.Sprintf("%s:%s", os.Getenv(constant.ENDPOINT), os.Getenv(constant.PLATINNERPORT))))
	result_tx := base64.URLEncoding.EncodeToString([]byte(txJsonString))
	return sendRequest(result_tx)
}

func GetFraBalance(from string) uint64 {
	return platform.GetUserFraBalance([]byte(from), []byte(fmt.Sprintf("%s:%s", os.Getenv(constant.ENDPOINT), os.Getenv(constant.PLATINNERPORT))))
}
