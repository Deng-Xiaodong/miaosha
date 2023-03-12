package rabbitmq

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

type RabbitMQ struct {
	Conn     *amqp.Connection
	Channel  *amqp.Channel
	Exchange string
	Queue    string
	BlindKey string
	MqUrl    string
}

var mqUrl string

func SetURL(url string) {
	mqUrl = url
}
func newRabbitMQ(que, eg, key string) *RabbitMQ {
	return &RabbitMQ{Exchange: eg, Queue: que, BlindKey: key, MqUrl: mqUrl}
}
func doFail(err error, msg string) {
	if err != nil {
		log.Fatalf(msg)
	}
}

func NewSimpleRabbitMQ(que string) *RabbitMQ {
	mq := newRabbitMQ(que, "", "")
	var err error
	mq.Conn, err = amqp.Dial(mq.MqUrl)
	doFail(err, fmt.Sprintf("%s:连接失败\n错误原因:%v", mq.MqUrl, err))
	mq.Channel, err = mq.Conn.Channel()
	doFail(err, "建立channel失败")
	return mq
}

func (mq *RabbitMQ) Publish(msg []byte) {
	_, err := mq.Channel.QueueDeclare(mq.Queue, false, false, false, false, nil)
	if err != nil {
		log.Println(err)
	}
	log.Println("生成一条信息")
	_ = mq.Channel.Publish(mq.Exchange, mq.Queue, false, false, amqp.Publishing{ContentType: "text/plain", Body: msg})
}

type Message struct {
	UserId string `json:"user_id"`
	ProdId string `json:"product_id"`
}

func (mq *RabbitMQ) Consumer() {
	q, _ := mq.Channel.QueueDeclare(mq.Queue, false, false, false, false, nil)
	msgs, err := mq.Channel.Consume(q.Name, "", false, false, false, false, nil)
	doFail(err, "消费失败")
	forever := make(chan struct{})
	go func() {
		println("我来消费了")
		for msg := range msgs {
			message := &Message{}
			err := json.Unmarshal(msg.Body, message)
			if err != nil {
				println(err)
			}
			log.Printf("用户：%s 抢购了产品%s", message.UserId, message.ProdId)
			_ = msg.Ack(false)
		}
	}()
	<-forever
}
