package db

import (
	"github/HyprNetwork/brc20-balance-monitor/constant"
	"os"
	"sync/atomic"

	libredis "github.com/redis/go-redis/v9"
)

var mredis struct {
	rdb atomic.Value
}

func MRedis() *libredis.Client {
	return mredis.rdb.Load().(*libredis.Client)
}

func init() {
	redisUri := os.Getenv(constant.REDISURL)
	opts, err := libredis.ParseURL(redisUri)
	if err != nil {
		panic(err)
	}

	r := libredis.NewClient(opts)

	mredis.rdb.Store(r)
}
