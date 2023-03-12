package rabbitmq

import (
	"encoding/json"
	"strconv"
	"sync"
	"testing"
)

func startPublish() {
	mq := NewSimpleRabbitMQ("dong")
	for i := 0; i < 10; i++ {
		msg, err := json.Marshal(&Message{UserId: strconv.Itoa(i), ProdId: strconv.Itoa(i)})
		if err != nil {
			println(err)
		}
		mq.Publish(msg)
	}

}
func startConsumer() {
	mq := NewSimpleRabbitMQ("dong")
	mq.Consumer()
}
func TestMQ(t *testing.T) {
	SetURL("amqp://guest:guest@miaosha.peadx.live:5672/miaosha")
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		startPublish()
	}()
	startConsumer()
	wg.Wait()

}
