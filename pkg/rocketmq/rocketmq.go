package rocketmq

import (
	"context"
	"fmt"
	"log"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
)

type Rocketmq struct {
	RocketmqProducer rocketmq.Producer
}

func InitRocketmqClient() *Rocketmq {
	p, err := rocketmq.NewProducer(
		producer.WithNameServer([]string{"8.156.86.220:9876"}),
		producer.WithRetry(2),
		producer.WithGroupName("PID_IM_SERVICE"),
	)
	if err != nil {
		panic(err)
	}
	err = p.Start()
	if err != nil {
		panic(err)
	}
	log.Println("RocketMQ Producer 启动成功")
	return &Rocketmq{RocketmqProducer: p}
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
