package handler

import (
	"github/HyprNetwork/brc20-balance-monitor/model"
	"github/HyprNetwork/brc20-balance-monitor/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetUserBalance(ctx *gin.Context) {
	b := &model.BRC20TokenBalance{}
	result, err := b.GetByTickerAndAddress(ctx.Query("ticker"), ctx.Query("address"))
	if err != nil {
		utils.GetLogger().Errorf("GetByTickerAndAddress err:%+v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func GetUserAllBalance(ctx *gin.Context) {
	b := &model.BRC20TokenBalance{}
	result, err := b.FindUserAllBalance(ctx.Query("address"))
	if err != nil {
		utils.GetLogger().Errorf("FindUserAllBalance err:%+v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func GetTokenList(ctx *gin.Context) {
	t := &model.Token{}
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
	findParams := model.FindParams{}
	if ctx.Query("type") == "" {
		findParams.Type = 0
	} else {
		typeFind, err := strconv.Atoi(ctx.Query("type"))
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
			return
		}
		findParams.Type = typeFind
	}
	findParams.Ticker = ctx.Query("ticker")

	result, err := t.FindPageList(pageNo, pageCount, findParams)
	if err != nil {
		utils.GetLogger().Errorf("FindPageList err:%+v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, result)
}

func GetTokenDetail(ctx *gin.Context) {
	idStr := ctx.Params.ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	t := &model.Token{}
	res, err := t.GetDetail(uint64(id))
	if err != nil {
		utils.GetLogger().Errorf("GetDetail err:%+v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, res)
}

func GetTokenUserRankList(ctx *gin.Context) {
	ticker := ctx.Query("ticker")
	pageNo, err := strconv.Atoi(ctx.Query("pageNo"))
	if err != nil {
		utils.GetLogger().Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	pageCount, err := strconv.Atoi(ctx.Query("pageCount"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	b := &model.BRC20TokenBalance{}
	res, err := b.GetUserList(ticker, pageNo, pageCount)
	if err != nil {
		utils.GetLogger().Errorf("GetUserList err:%+v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, res)
}

func CheckTokenExist(ctx *gin.Context) {
	ticker := ctx.Params.ByName("ticker")

	token := &model.Token{}
	result, err := token.CheckTicker(ticker)
	if err != nil {
		utils.GetLogger().Errorf("CheckTicker err:%+v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, result)
}
