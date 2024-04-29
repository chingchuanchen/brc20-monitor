package handler

import (
	"bytes"
	"fmt"
	"github/HyprNetwork/brc20-balance-monitor/constant"
	"github/HyprNetwork/brc20-balance-monitor/utils"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	libredis "github.com/redis/go-redis/v9"

	limiter "github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	sredis "github.com/ulule/limiter/v3/drivers/store/redis"
)

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
func (w bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func LoggerToFile() gin.HandlerFunc {
	return func(c *gin.Context) {

		bodyLogWriter := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = bodyLogWriter
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()

		// 执行时间
		latencyTime := endTime.Sub(startTime)

		// 请求方式
		reqMethod := c.Request.Method

		// 请求路由
		reqUri := c.Request.RequestURI

		// 状态码
		statusCode := c.Writer.Status()

		// 请求IP
		clientIP := c.ClientIP()

		// 日志格式
		if statusCode == 200 {
			utils.GetLogger().Infof("| %3d | %13v | %15s | %s | %s",
				statusCode,
				latencyTime,
				clientIP,
				reqMethod,
				reqUri,
			)
		} else {
			utils.GetLogger().Errorf("| %3d | %13v | %15s | %s | %s | %s",
				statusCode,
				latencyTime,
				clientIP,
				reqMethod,
				reqUri,
				bodyLogWriter.body.String(),
			)
		}

	}
}

func RateLimitSecond() gin.HandlerFunc {
	// Define a limit rate requests per second.
	rate, err := limiter.NewRateFromFormatted(fmt.Sprintf("%s-S", os.Getenv(constant.RATELIMITSECOND)))
	if err != nil {
		log.Fatal(err)
	}

	// Create a redis client.
	option, err := libredis.ParseURL(os.Getenv(constant.REDISURL))
	if err != nil {
		log.Fatal(err)
	}
	client := libredis.NewClient(option)

	// Create a store with the redis client.
	store, err := sredis.NewStoreWithOptions(client, limiter.StoreOptions{
		Prefix:   "limiter_second",
		MaxRetry: 3,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create a new middleware with the limiter instance.
	return mgin.NewMiddleware(limiter.New(store, rate))

}

func RateLimitDay() gin.HandlerFunc {
	// Define a limit rate requests per day.
	rate, err := limiter.NewRateFromFormatted(fmt.Sprintf("%s-D", os.Getenv(constant.RATELIMITDAY)))
	if err != nil {
		log.Fatal(err)
	}

	// Create a redis client.
	option, err := libredis.ParseURL(os.Getenv(constant.REDISURL))
	if err != nil {
		log.Fatal(err)
	}
	client := libredis.NewClient(option)

	// Create a store with the redis client.
	store, err := sredis.NewStoreWithOptions(client, limiter.StoreOptions{
		Prefix:   "limiter_day",
		MaxRetry: 3,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create a new middleware with the limiter instance.
	return mgin.NewMiddleware(limiter.New(store, rate))
}
