package process

import (
	"Hyper/pkg/log"
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	rmq_client "github.com/apache/rocketmq-clients/golang/v5"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

var (
	// maximum waiting time for receive func
	awaitDuration = time.Second * 5
	// maximum number of messages received at one time
	maxMessageNum int32 = 16
	// invisibleDuration should > 20s
	invisibleDuration = time.Second * 20
	// receive concurrency
	receiveConcurrency = 10
)

var once sync.Once

type IServer interface {
	Setup(ctx context.Context) error
	Init() error
}

// SubServers 订阅的服务列表
type SubServers struct {
	HealthSubscribe  *HealthSubscribe  // 注册健康上报
	MessageSubscribe *MessageSubscribe /// 注册消息订阅
	NoticeSubscribe  *NoticeSubscribe
}

type Server struct {
	items      []IServer
	MqConsumer rmq_client.SimpleConsumer
	SubServers
}

func NewServer(servers *SubServers, mqConsumer rmq_client.SimpleConsumer) *Server {
	s := &Server{
		MqConsumer: mqConsumer,
		SubServers: *servers,
	}

	s.binds(servers)
	return s
}

func (c *Server) binds(servers *SubServers) {
	elem := reflect.ValueOf(servers).Elem()
	for i := 0; i < elem.NumField(); i++ {
		if v, ok := elem.Field(i).Interface().(IServer); ok {
			c.items = append(c.items, v)
		}
	}
}

// Start 启动服务
func (c *Server) Start(eg *errgroup.Group, ctx context.Context) {
	once.Do(func() {
		for _, process := range c.items {
			if err := process.Init(); err != nil {
				log.L.Fatal("注册 Topic 失败", zap.Error(err))
			}
		}

		for _, process := range c.items {
			serv := process
			eg.Go(func() error {
				return serv.Setup(ctx)
			})
		}

		if err := c.MqConsumer.Start(); err != nil {
			log.L.Fatal("Failed to start consumer", zap.Error(err))
		}

		eg.Go(func() error {
			<-ctx.Done()
			log.L.Info("正在优雅关闭 RocketMQ 消费者...")
			return c.MqConsumer.GracefulStop()
		})

		log.L.Info("start receive message")

		// 5. 启动并发消费
		for i := 0; i < receiveConcurrency; i++ {
			eg.Go(func() error {
				for {
					select {
					case <-ctx.Done():
						return nil
					default:
						mvs, _ := c.MqConsumer.Receive(ctx, maxMessageNum, invisibleDuration)
						for _, mv := range mvs {
							if mv == nil {
								continue
							}
							if err := c.processMessage(ctx, mv); err != nil {
								// 处理失败：不要 Ack，让 MQ 在 invisibleDuration 之后重投
								continue
							}
							// 处理成功：Ack
							if err := c.MqConsumer.Ack(ctx, mv); err != nil {
								log.L.Error("ack message error", zap.Error(err))
							}
						}
					}
				}
			})
		}
	})
}

// 建议提取一个简单的处理函数，保持代码整洁
func (c *Server) processMessage(ctx context.Context, mv *rmq_client.MessageView) error {

	topic := mv.GetTopic()
	var err error

	switch topic {
	case "IM_CHAT_MSGS":
		if c.MessageSubscribe != nil {
			err = c.MessageSubscribe.handleMessage(ctx, mv)
		}
	case "HYPER_SYSTEM_MSGS":
		if c.NoticeSubscribe != nil {
			_, err = c.NoticeSubscribe.handleSystem(ctx, mv)
		}
	default:
		// 不认识的 topic，直接返回错误看日志
		err = fmt.Errorf("unknown topic: %s", topic)
	}

	if err != nil {
		log.L.Warn("消息处理失败", zap.String("topic", topic), zap.Error(err))
		return err
	}
	return nil
}
