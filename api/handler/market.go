package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github/HyprNetwork/brc20-balance-monitor/constant"
	"github/HyprNetwork/brc20-balance-monitor/db"
	"github/HyprNetwork/brc20-balance-monitor/model"
	"github/HyprNetwork/brc20-balance-monitor/platform"
	"github/HyprNetwork/brc20-balance-monitor/service"
	"github/HyprNetwork/brc20-balance-monitor/utils"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func AddList(ctx *gin.Context) {
	// 上架信息
	ticker := ctx.PostForm("ticker")
	user := ctx.PostForm("user")
	price := ctx.PostForm("totalPrice")
	if price == "" {
		price = "0"
	}
	amount := ctx.PostForm("amount")
	if amount == "" {
		amount = "0"
	}
	listRecord := &model.ListRecord{
		Ticker:         ticker,
		User:           user,
		Price:          price,
		Amount:         amount,
		CenterMnemonic: platform.GetMnemonic(),
		State:          constant.ListWaiting,
	}
	lastInsertId, err := listRecord.InsertToDB()
	if err != nil {
		utils.GetLogger().Errorf("InsertToDB err:%+v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"result": "ok", "listId": lastInsertId})
}

func ConfirmList(ctx *gin.Context) {
	// 确认上架
	idStr := ctx.PostForm("listId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	user := ctx.PostForm("user")
	listRecord := &model.ListRecord{Base: model.Base{Id: uint64(id)}, User: user}
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	record, err := listRecord.GetById(id)
	if err != nil {
		utils.GetLogger().Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	if record.User != user {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.New("invalid user").Error())
		return
	}

	err = listRecord.ConfirmList()
	if err != nil {
		utils.GetLogger().Errorf("ConfirmList err:%+v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"result": "ok"})
}

func CancelList(ctx *gin.Context) {
	// 取消上架
	idStr := ctx.PostForm("listId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	user := ctx.PostForm("user")
	listRecord := &model.ListRecord{Base: model.Base{Id: uint64(id)}, User: user}
	tx, err := db.Master().Begin()
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	err = listRecord.Cancel()
	if err != nil {
		tx.Rollback()
		utils.GetLogger().Errorf("Cancel err:%+v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	// 需要中心化账户把钱转回去
	record, err := listRecord.GetById(id)
	if err != nil {
		utils.GetLogger().Error(err)
		tx.Rollback()
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	if record.User != user {
		tx.Rollback()
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.New("invalid user").Error())
		return
	}
	toPubkey, err := utils.GetPubkeyFromAddress(record.User)
	if err != nil {
		tx.Rollback()
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	_, err = utils.SendTx(record.CenterMnemonic, toPubkey, toPubkey, record.Amount, record.Ticker, record.Price, constant.BRC20_OP_TRANSFER)
	if err != nil {
		tx.Rollback()
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	if err := tx.Commit(); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"result": "ok"})
}

func CenterAccount(ctx *gin.Context) {
	idStr := ctx.Query("listId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	listRecord := &model.ListRecord{Base: model.Base{Id: uint64(id)}}
	record, err := listRecord.GetById(id)
	if err != nil {
		utils.GetLogger().Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"result": platform.Mnemonic2Bench32([]byte(record.CenterMnemonic))})
}

func ConfirmBuyAndCheck(ctx *gin.Context) {
	// 挂单id
	idStr := ctx.PostForm("listId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}
	// 购买者
	user := ctx.PostForm("user")
	// 购买者向中心化账户转账的交易hash
	txCheck := ctx.PostForm("tx")
	listRecord := &model.ListRecord{Base: model.Base{Id: uint64(id)}, ToUser: user}
	tx, err := db.Master().Begin()
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	record, err := listRecord.GetById(id)
	if err != nil {
		utils.GetLogger().Error(err)
		tx.Rollback()
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	reqUrl := fmt.Sprintf("%s:%s/tx?hash=0x%s", os.Getenv(constant.ENDPOINT), os.Getenv(constant.PLATAPIPORT), txCheck)
	trans := &service.Transaction{}

	// 检查购买者是否已经给中心化账户转fra
	center, err := utils.GetPubkeyFromAddress(platform.Mnemonic2Bench32([]byte(record.CenterMnemonic)))
	if err != nil {
		utils.GetLogger().Errorf("GetPubkeyFromAddress err:%+v", err)
		tx.Rollback()
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	_, err = trans.CheckBuy(reqUrl, user, center, record.Price)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	// 如果转好了，则进行下面流程
	err = listRecord.Finished()
	if err != nil {
		utils.GetLogger().Errorf("Finished err:%+v", err)
		tx.Rollback()
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	// 需要中心化账户把brc20 token打给购买者, 并且将fra转给上架者
	toPubkey, err := utils.GetPubkeyFromAddress(user)
	if err != nil {
		tx.Rollback()
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	receiver, err := utils.GetPubkeyFromAddress(record.User)
	if err != nil {
		tx.Rollback()
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	_, err = utils.SendTx(record.CenterMnemonic, receiver, toPubkey, record.Amount, record.Ticker, record.Price, constant.BRC20_OP_TRANSFER)
	if err != nil {
		utils.GetLogger().Errorf("SendTx err:%+v", err)
		tx.Rollback()
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	if err := tx.Commit(); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	// 购买之后可能会导致floor price 更新
	go func(ch chan<- *model.TokenMarketInfo, id int) {
		listRecord := &model.ListRecord{}
		info, _ := listRecord.GetMarketInfo(id)
		b := &model.BRC20TokenBalance{}
		hodler, _ := b.GetTickerHolders(info.Ticker)
		ch <- &model.TokenMarketInfo{
			Ticker:      info.Ticker,
			Holders:     hodler,
			FloorPrice:  info.FloorPrice,
			TotalVal24h: info.TotalVal24h,
			TotalVal:    info.TotalVal,
		}

	}(globalInfoChangeChannel, id)

	ctx.JSON(http.StatusOK, gin.H{"result": "ok"})
}

func GetListRecords(ctx *gin.Context) {
	pageNo, err := strconv.Atoi(ctx.Query("pageNo"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	pageCount, err := strconv.Atoi(ctx.Query("pageCount"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	findParams := model.UserTickerListFindParams{}
	if ctx.Query("state") == "" {
		findParams.State = 0
	} else {
		state, err := strconv.Atoi(ctx.Query("state"))
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
			return
		}
		findParams.State = state
	}
	ticker := ctx.Query("ticker")
	user := ctx.Query("user")
	findParams.Ticker = ticker
	findParams.User = user

	record := &model.ListRecord{}
	result, err := record.FindPageList(pageNo, pageCount, findParams)
	if err != nil {
		utils.GetLogger().Errorf("FindPageList err:%+v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func GetOrderListRecords(ctx *gin.Context) {
	pageNo, err := strconv.Atoi(ctx.Query("pageNo"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}
	pageCount, err := strconv.Atoi(ctx.Query("pageCount"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	findParams := model.UserTickerListFindParams{}
	ticker := ctx.Query("ticker")
	user := ctx.Query("user")
	findParams.Ticker = ticker
	findParams.User = user

	record := &model.ListRecord{}
	result, err := record.FindOrderPageList(pageNo, pageCount, findParams)
	if err != nil {
		utils.GetLogger().Errorf("FindPageList err:%+v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, result)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	globalInfoChangeChannel = make(chan *model.TokenMarketInfo)
)

func UpdateMarketInfo(ctx *gin.Context) {
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	defer conn.Close()

	for {
		select {
		case marketInfo := <-globalInfoChangeChannel:
			data, err := json.Marshal(marketInfo)
			if err != nil {
				utils.GetLogger().Errorf("serializing market info: %v", err)
				continue
			}
			err = conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				utils.GetLogger().Errorf("WriteMessage market info: %v", err)
				continue
			}
		case <-time.After(time.Second):
			continue
		}
	}
}

func GetMarketIndex(ctx *gin.Context) {
	ticker := ctx.Query("ticker")
	param := model.MarketSearchParam{
		Ticker: "%" + ticker,
	}
	pageNo, err := strconv.Atoi(ctx.Query("pageNo"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	pageCount, err := strconv.Atoi(ctx.Query("pageCount"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	token := &model.Token{}
	res, err := token.FindMarketInfos(pageNo, pageCount, param)
	if err != nil {
		utils.GetLogger().Errorf("FindMarketInfos err:%+v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func GetBanner(ctx *gin.Context) {
	param := model.MarketSearchParam{}
	token := &model.Token{}
	res, err := token.FindMarketInfos(1, 5, param)
	if err != nil {
		utils.GetLogger().Errorf("FindMarketInfos err:%+v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, res.Data)
}
