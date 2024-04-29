package service

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github/HyprNetwork/brc20-balance-monitor/constant"
	"github/HyprNetwork/brc20-balance-monitor/db"
	"github/HyprNetwork/brc20-balance-monitor/decimal"
	"github/HyprNetwork/brc20-balance-monitor/model"
	"github/HyprNetwork/brc20-balance-monitor/utils"
	"strings"
	"time"
)

type TxHandler struct {
	block model.BlockInfo
}

type TransferListEntity struct {
	OutputPubKey string
	Amount       string
}

func NewTxHandler(block model.BlockInfo) *TxHandler {
	return &TxHandler{block: block}
}

func (t *TxHandler) Handle(height int64) error {

	for _, v := range t.block.Data.Txs {
		txBytes, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			utils.GetLogger().Errorf("decode tx err : %+v", err)
			return err
		}
		if strings.Contains(string(txBytes), "evm") {
			utils.GetLogger().Warn("evm tx")
			continue
		}
		var operations model.HanldeBody
		err = json.Unmarshal(txBytes, &operations)
		if err != nil {
			utils.GetLogger().Warn(string(txBytes))
			utils.GetLogger().Errorf("unmarshal operations err : %+v", err)
			return err
		}
		for _, op := range operations.Body.Operations {
			if len(op.TransferAsset.Body.Outputs) == 0 {
				continue
			}
			tranferTo := make(map[string][]TransferListEntity, 0)
			for _, output := range op.TransferAsset.Body.Outputs {
				if output.Memo == "" {
					continue
				}
				tempMemo := new(model.InscriptionBRC20)
				err = tempMemo.Unmarshal([]byte(output.Memo))
				if err != nil {
					utils.GetLogger().Warnf("memo invalid %s", output.Memo)
					continue
				}

				switch tempMemo.Operation {
				case constant.BRC20_OP_DEPLOY:
					err := t.HandleDeploy(output.Record.PublicKey, tempMemo)
					if err != nil {
						return err
					}
				case constant.BRC20_OP_TRANSFER:
					if vlist, ok := tranferTo[tempMemo.BRC20Tick]; ok {
						vlist = append(vlist, TransferListEntity{
							Amount:       tempMemo.BRC20Amount,
							OutputPubKey: output.Record.PublicKey,
						})
						tranferTo[tempMemo.BRC20Tick] = vlist
					} else {
						tranferTo[tempMemo.BRC20Tick] = make([]TransferListEntity, 0)
						tranferTo[tempMemo.BRC20Tick] = append(tranferTo[tempMemo.BRC20Tick], TransferListEntity{
							Amount:       tempMemo.BRC20Amount,
							OutputPubKey: output.Record.PublicKey,
						})
					}
				case constant.BRC20_OP_MINT:
					err = t.HandleMint(output.Record.PublicKey, tempMemo, height)
					if err != nil {
						return err
					}
				default:
					return errors.New("unsupported operation")
				}
			}

			if len(tranferTo) != 0 {
				err = t.HandleTransfer(op.TransferAsset.Body.Transfer.Inputs, tranferTo, height)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (t *TxHandler) HandleDeploy(pubKey string, script *model.InscriptionBRC20) error {
	from, err := utils.GetAddressFromScript(pubKey)
	if err != nil {
		return err
	}
	h := NewHandler(from, "", script)
	return h.HandleDepoly()
}

func (t *TxHandler) HandleMint(pubKey string, script *model.InscriptionBRC20, height int64) error {
	utils.GetLogger().Infof("HandleMint %v", script)
	from, err := utils.GetAddressFromScript(pubKey)
	if err != nil {
		return err
	}
	h := NewHandler(from, "", script)
	return h.HandleMint(height)
}

func (t *TxHandler) HandleTransfer(senders []model.Input, receiverMap map[string][]TransferListEntity, height int64) error {
	for k, v := range receiverMap {
		err := t.handleTransfer(senders, k, v, height)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *TxHandler) handleTransfer(senders []model.Input, ticker string, receivers []TransferListEntity, height int64) error {
	lastedBlock, err := model.GetLatestBlock()
	if err != nil {
		return err
	}
	var (
		amountTotal  = decimal.NewDecimal()
		receiveTotal = decimal.NewDecimal()
	)
	// check token
	uniqueLowerTicker := strings.ToLower(ticker)
	tokenInfo, err := model.NewTokenFromDBByTicker(uniqueLowerTicker)
	if err != nil {
		return err
	}
	if tokenInfo.IsEmpty() {
		return fmt.Errorf("handle transfer but token not existed. ticker: %s", uniqueLowerTicker)
	}

	tokenInfoMax, err := tokenInfo.GetMax()
	if err != nil {
		return err
	}
	// check inputs
	inputs := make([]*model.BRC20TokenBalance, 0)
	inputCounted := make(map[string]bool, 0)
	for _, v := range senders {
		account, err := utils.GetAddressFromScript(v.PublicKey)
		if err != nil {
			return err
		}
		// 去重
		if _, ok := inputCounted[account]; ok {
			continue
		} else {
			inputCounted[account] = true
		}
		inputTemp, err := model.NewBRC20TokenBalanceFromDB(ticker, account)
		if err != nil {
			return err
		}
		if inputTemp.IsEmpty() {
			utils.GetLogger().Warnf("transfer account %s not exist for ticker: %s", account, ticker)
		}
		// 该用户在后面区块已经计算过了，不用再同步
		if inputTemp.Height > lastedBlock.Height {
			return nil
		}
		// 当前区块已经同步完成
		if inputTemp.Height == lastedBlock.Height && lastedBlock.IsSyncFinished() {
			return nil
		}
		originAmount, err := inputTemp.GetOverallBalance()
		if err != nil {
			return err
		}
		// 当前ticker下的balance全部用掉，等待下次找零的恢复
		inputTemp.OverallBalance = "0"
		inputs = append(inputs, inputTemp)
		amountTotal = amountTotal.Add(originAmount)
	}
	if amountTotal.Sign() <= 0 || amountTotal.Cmp(tokenInfoMax) > 0 {
		return errors.New("transfer token amount is invalid")
	}

	outputs := make([]*model.BRC20TokenBalance, 0)
	for _, v := range receivers {
		account, err := utils.GetAddressFromScript(v.OutputPubKey)
		if err != nil {
			return err
		}
		outputTemp, err := model.NewBRC20TokenBalanceFromDB(ticker, account)
		if err != nil {
			return err
		}
		curAmount, precision, err := decimal.NewDecimalFromString(v.Amount)
		if err != nil {
			return err
		}
		if precision > int(tokenInfo.Decimal) {
			return errors.New("receive token is too small")
		}
		originAmount, err := outputTemp.GetOverallBalance()
		if err != nil {
			return err
		}
		if inputCounted[account] {
			outputTemp.OverallBalance = curAmount.String()
		} else {
			outputTemp.OverallBalance = originAmount.Add(curAmount).String()
		}

		outputs = append(outputs, outputTemp)
		receiveTotal = receiveTotal.Add(curAmount)
	}
	if receiveTotal.Sign() <= 0 || receiveTotal.Cmp(tokenInfoMax) > 0 || receiveTotal.Cmp(amountTotal) > 0 {
		return errors.New("receive token amount is invalid")
	}

	return t.handleTransferInSession(inputs, outputs, height)

}

func (t *TxHandler) handleTransferInSession(inputs, outputs []*model.BRC20TokenBalance, height int64) error {
	tx := db.Master().MustBegin()

	// from to dec amount
	for _, v := range inputs {
		tx.MustExec("update balance set overall_balance = $1, update_time = $2, height = $3 where ticker = $4 and address = $5", v.OverallBalance, time.Now().Unix(), height, v.Ticker, v.Address)
	}

	// receive to add amount
	for _, v := range outputs {
		if v.IsEmpty() {
			tx.MustExec("INSERT INTO balance (address, ticker, overall_balance, create_time, update_time, height) values ($1, $2, $3, $4, $5, $6)", v.Address, v.Ticker, v.OverallBalance, time.Now().Unix(), time.Now().Unix(), height)
		} else {
			tx.MustExec("update balance set overall_balance = $1, update_time = $2, height = $3 where ticker = $4 and address = $5", v.OverallBalance, time.Now().Unix(), height, v.Ticker, v.Address)
		}
	}

	return tx.Commit()
}
