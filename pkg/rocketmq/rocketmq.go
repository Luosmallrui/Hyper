package rocketmq

import (
	"context"
	"fmt"
	"log"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/apache/rocketmq-client-go/v2/rlog"
)

type Rocketmq struct {
	RocketmqProducer rocketmq.Producer
	RocketmqConsumer rocketmq.PushConsumer
}

func init() {
	initRocketMQLogger()
}

func InitProducer() rocketmq.Producer {
	p, err := rocketmq.NewProducer(
		producer.WithNameServer([]string{"8.156.86.220:9876"}),
		producer.WithRetry(2),
		producer.WithGroupName("PID_IM_SERVICE"),
	)
	if err != nil {
		panic(err)
	}

	if err = p.Start(); err != nil {
		return nil
	}

	log.Println("RocketMQ Producer 启动成功")
	return p
}
func InitConsumer() rocketmq.PushConsumer {
	c, err := rocketmq.NewPushConsumer(
		consumer.WithNameServer([]string{"8.156.86.220:9876"}),
		consumer.WithGroupName("IM_STORAGE_GROUP"),
	)
	if err != nil {
		panic(err)
	}

	return c
}

func (p *Rocketmq) SendMsg(topic string, body []byte) error {
	msg := &primitive.Message{
		Topic: topic,
		Body:  body,
	}

	// 发送同步消息
	res, err := p.RocketmqProducer.SendSync(context.Background(), msg)
	if err != nil {
		return err
	}
	fmt.Printf("发送成功: %s \n", res.MsgID)
	return nil
}

func initRocketMQLogger() {
	rlog.SetLogLevel("info")
	rlog.SetOutputPath("/root/logs/rocketmq.log")
}
