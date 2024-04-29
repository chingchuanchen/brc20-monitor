package main

import (
	"flag"
	"fmt"
	"github/HyprNetwork/brc20-balance-monitor/model"
	"github/HyprNetwork/brc20-balance-monitor/utils"
	"net/http"

	"github/HyprNetwork/brc20-balance-monitor/api/handler"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

func GetLatestHeight(ctx *gin.Context) {
	b := &model.Block{}
	result, err := b.LatestHeight()
	if err != nil {
		utils.GetLogger().Errorf("LatestHeight err:%+v", err)
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"result": result})
}

func main() {
	var port = flag.String("port", "8090", "listen port")
	flag.Parse()

	r := gin.New()
	r.Use(handler.Cors())

	r.Use(handler.LoggerToFile())

	// token.go
	r.GET("/balance", handler.GetUserBalance)
	r.GET("/balance/all", handler.GetUserAllBalance)
	r.GET("/tokenList", handler.GetTokenList)
	r.GET("/token/:id/detail", handler.GetTokenDetail)
	r.GET("/token/userRank", handler.GetTokenUserRankList)
	r.GET("/token/check/:ticker", handler.CheckTokenExist)

	// market.go
	r.POST("/addList", handler.AddList)
	r.POST("/confirmList", handler.ConfirmList)
	r.GET("/account", handler.CenterAccount)
	r.POST("/cancelList", handler.CancelList)
	r.POST("/buy", handler.ConfirmBuyAndCheck)
	r.GET("/list", handler.GetListRecords)
	r.GET("/orderList", handler.GetOrderListRecords)
	r.GET("/myList", handler.GetOrderListRecords)
	r.GET("/market", handler.GetMarketIndex)
	r.GET("/banner", handler.GetBanner)

	r.GET("/marketInfoWS", handler.UpdateMarketInfo)

	// airdrop.go
	r.POST("/airdrop", handler.RateLimitSecond(), handler.RateLimitDay(), handler.AirDropCheckAndSend)

	r.GET("/height", GetLatestHeight)
	pprof.Register(r)
	r.Run(fmt.Sprintf(":%s", *port))
}
