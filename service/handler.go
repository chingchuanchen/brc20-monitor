package service

import (
	"errors"
	"fmt"
	"github/HyprNetwork/brc20-balance-monitor/db"
	"github/HyprNetwork/brc20-balance-monitor/decimal"
	"github/HyprNetwork/brc20-balance-monitor/model"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	From   string
	To     string
	Script *model.InscriptionBRC20
}

func NewHandler(from, to string, script *model.InscriptionBRC20) *Handler {
	return &Handler{From: from, To: to, Script: script}
}

// HandleDeploy deploy inscription
func (h *Handler) HandleDepoly() error {
	uniqueLowerTicker := strings.ToLower(h.Script.BRC20Tick)
	if h.Script.BRC20Max == "" {
		return fmt.Errorf("handle deploy but max missing. ticker: %s", uniqueLowerTicker)
	}
	tokenInfo := &model.Token{Ticker: uniqueLowerTicker, DeployUser: h.From}
	if dec, err := strconv.ParseUint(h.Script.BRC20Decimal, 10, 64); err != nil {
		return err
	} else {
		tokenInfo.Decimal = uint8(dec)
	}

	if max, precision, err := decimal.NewDecimalFromString(h.Script.BRC20Max); err != nil {
		return err
	} else {
		if max.Sign() <= 0 || max.IsOverflowUint64() || precision > int(tokenInfo.Decimal) {
			return errors.New("invalid max")
		}
		tokenInfo.Max = max.String()
	}

	if lim, precision, err := decimal.NewDecimalFromString(h.Script.BRC20Limit); err != nil {
		return err
	} else {
		if lim.Sign() <= 0 || lim.IsOverflowUint64() || precision > int(tokenInfo.Decimal) {
			return errors.New("invalid limit")
		}
		tokenInfo.Limit = lim.String()
	}

	return tokenInfo.InsertToDB()
}

// HandleTransfer handle transfer script
func (h *Handler) HandleTransfer() error {
	lastedBlock, err := model.GetLatestBlock()
	if err != nil {
		return err
	}
	uniqueLowerTicker := strings.ToLower(h.Script.BRC20Tick)
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

	// check transfer amount
	amt, precision, err := decimal.NewDecimalFromString(h.Script.BRC20Amount)
	if err != nil {
		return err
	}
	if precision > int(tokenInfo.Decimal) {
		return errors.New("transfer token is too small")
	}
	if amt.Sign() <= 0 || amt.Cmp(tokenInfoMax) > 0 {
		return errors.New("transfer token amount is invalid")
	}

	fromRecord, err := model.NewBRC20TokenBalanceFromDB(h.Script.BRC20Tick, h.From)
	if err != nil {
		return err
	}
	if fromRecord.IsEmpty() {
		return errors.New("transfer from must exist")
	}
	// 该用户在后面区块已经计算过了，不用再同步
	if fromRecord.Height > lastedBlock.Height {
		return nil
	}
	// 当前区块已经同步完成
	if fromRecord.Height == lastedBlock.Height && lastedBlock.IsSyncFinished() {
		return nil
	}
	fromRecordBalance, err := fromRecord.GetOverallBalance()
	if err != nil {
		return err
	}
	if fromRecordBalance.Cmp(amt) < 0 {
		return errors.New("transfer from must have enough balance")
	}

	fromRecord.OverallBalance = fromRecordBalance.Sub(amt).String()

	toRecord, err := model.NewBRC20TokenBalanceFromDB(h.Script.BRC20Tick, h.To)
	if err != nil {
		return err
	}
	toRecordBalance, err := toRecord.GetOverallBalance()
	if err != nil {
		return err
	}
	toRecord.OverallBalance = toRecordBalance.Add(amt).String()

	return h.handleTransferInSession(fromRecord, toRecord)
}

func (h *Handler) handleTransferInSession(fromRecord, toRecord *model.BRC20TokenBalance) error {
	tx := db.Master().MustBegin()

	// 1. from update
	tx.MustExec("update balance set overall_balance = $1, update_time = $2 where id = $3", fromRecord.OverallBalance, time.Now().Unix(), fromRecord.Id)
	// 2. to update or insert
	if toRecord.IsEmpty() {
		tx.MustExec("INSERT INTO balance (address, ticker, overall_balance, create_time, update_time) values ($1, $2, $3, $4, $5)", toRecord.Address, toRecord.Ticker, toRecord.OverallBalance, time.Now().Unix(), time.Now().Unix())
	} else {
		tx.MustExec("update balance set overall_balance = $1, update_time = $2 where id = $3", toRecord.OverallBalance, time.Now().Unix(), toRecord.Id)
	}

	return tx.Commit()
}

// HandleMint handle mint script
func (h *Handler) HandleMint(height int64) error {
	uniqueLowerTicker := strings.ToLower(h.Script.BRC20Tick)
	tokenInfo, err := model.NewTokenFromDBByTicker(uniqueLowerTicker)
	if err != nil {
		return err
	}

	if tokenInfo.IsEmpty() {
		return fmt.Errorf("handle mint but token not existed. ticker: %s", uniqueLowerTicker)
	}

	amt, precision, err := decimal.NewDecimalFromString(h.Script.BRC20Amount)
	if err != nil {
		return err
	}
	if precision > int(tokenInfo.Decimal) {
		return errors.New("mint token is too small")
	}

	m := &model.MintRecord{}
	mintTotal, err := m.MintTickerTotal(tokenInfo.Ticker)
	if err != nil {
		return err
	}

	mintTotalDecimal, _, err := decimal.NewDecimalFromString(mintTotal)
	if err != nil {
		return err
	}

	tokenMax, err := tokenInfo.GetMax()
	if err != nil {
		return err
	}

	tokenInfoLimit, err := tokenInfo.GetLimit()
	if err != nil {
		return err
	}

	// check mint amount
	if amt.Sign() <= 0 || amt.Cmp(tokenInfoLimit) > 0 || amt.Add(mintTotalDecimal).Cmp(tokenMax) > 0 {
		return errors.New("mint token amount is invalid")
	}

	record, err := model.NewBRC20TokenBalanceFromDB(h.Script.BRC20Tick, h.From)
	if err != nil {
		return err
	}

	recordBalance, err := record.GetOverallBalance()
	if err != nil {
		return err
	}

	record.Height = height

	record.OverallBalance = recordBalance.Add(amt).String()
	return h.handleMintInSession(record, amt)
}

func (h *Handler) handleMintInSession(record *model.BRC20TokenBalance, amount *decimal.Decimal) error {
	tx := db.Master().MustBegin()
	if record.IsEmpty() {
		err := record.InsertToDB()
		if err != nil {
			return tx.Rollback()
		}
	} else {
		err := record.UpdateToDB()
		if err != nil {
			return tx.Rollback()
		}
	}
	m := &model.MintRecord{Ticker: record.Ticker, User: record.Address, Amount: amount.String()}
	err := m.InsertToDB()
	if err != nil {
		return tx.Rollback()
	}

	return tx.Commit()
}
