package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log"
	"miaosha/common"
	"miaosha/rabbitmq"
	"miaosha/redislock"
	"os"
)

func main() {

	//初始化配置
	var configFile string
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}
	config := common.InitConfig(configFile)

	//初始化redis客户端连接
	redislock.InitRedisClient(config)

	//初始化RabbitMQ连接
	rabbitmq.SetURL(config.MqCfg.MqUrl)
	mq := rabbitmq.NewSimpleRabbitMQ(config.MqCfg.QueName)

	engine := gin.Default()
	engine.GET("/getone", func(ctx *gin.Context) {
		if v := redislock.GetLock(); v > 0 {
			uid := ctx.Param("user_id")
			pid := ctx.Param("product_id")
			msg, err := json.Marshal(rabbitmq.Message{UserId: uid, ProdId: pid})
			if err != nil {
				log.Fatal(err)
			}
			mq.Publish(msg)
			ctx.JSON(200, common.Error{Code: 200, Msg: "抢购成功"})
		} else {
			ctx.JSON(200, common.Error{Code: 500, Msg: "抢购失败"})
		}
	})

	_ = engine.Run(":" + config.GinCfg.Port)
}
