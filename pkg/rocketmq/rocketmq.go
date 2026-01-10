package rocketmq

import (
	"Hyper/config"
	"Hyper/pkg/log"
	"context"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/apache/rocketmq-client-go/v2/rlog"
	"go.uber.org/zap"
)

type Rocketmq struct {
	RocketmqProducer rocketmq.Producer
	RocketmqConsumer rocketmq.PushConsumer
}

func init() {
	rlog.SetLogLevel("error")
	//rlog.SetOutputPath("/Users/luosmallrui/Downloads/22583504_hypercn.cn_other/rocketmq.log")
}

func InitProducer(cfg *config.RocketMQConfig) rocketmq.Producer {
	p, err := rocketmq.NewProducer(
		producer.WithNameServer(cfg.NameServer),
		producer.WithRetry(cfg.Producer.Retry),
		producer.WithGroupName(cfg.Producer.Group),
	)
	if err != nil {
		panic(err)
	}

	if err = p.Start(); err != nil {
		return nil
	}
	log.L.Info("init producer success")

	return p
}
func InitConsumer(cfg *config.RocketMQConfig) rocketmq.PushConsumer {
	c, err := rocketmq.NewPushConsumer(
		consumer.WithNameServer(cfg.NameServer),
		consumer.WithGroupName(cfg.Consumer.Group),
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
	log.L.Info("seed message success", zap.Any("msg", res.MsgID))
	return nil
}
