package main

import (
	"miaosha/common"
	"miaosha/rabbitmq"
	"os"
)

func main() {
	//初始化配置
	var configFile string
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}
	config := common.InitMqConfig(configFile)
	//初始化RabbitMQ连接
	mq := rabbitmq.NewSimpleRabbitMQ(config.QueName, config.MqUrl)
	mq.Consumer()
}
