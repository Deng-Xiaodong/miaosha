package main

import "miaosha/rabbitmq"

const queName = "peadx"

func main() {
	mq := rabbitmq.NewSimpleRabbitMQ(queName)
	mq.Consumer()
}
