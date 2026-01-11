package process

import (
	"Hyper/pkg/log"
	"context"
	"reflect"
	"sync"

	"github.com/apache/rocketmq-client-go/v2"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
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
	mqConsumer rocketmq.PushConsumer
}

func NewServer(servers *SubServers, mqConsumer rocketmq.PushConsumer) *Server {
	s := &Server{mqConsumer: mqConsumer}

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
		// 1. 同步注册：一个一个来，绝对不会并发写 map
		for _, process := range c.items {
			if err := process.Init(); err != nil {
				log.L.Fatal("注册 Topic 失败", zap.Error(err))
			}
		}

		// 2. 统一启动：所有 Topic 注册完成后，调一次 Start 即可
		if err := c.mqConsumer.Start(); err != nil {
			log.L.Fatal("MQ 启动失败", zap.Error(err))
		}
		log.L.Info("RocketMQ 消费者已统一启动")

		// 3. 并发运行其他 Setup 逻辑（如健康检查的定时器等）
		for _, process := range c.items {
			serv := process
			eg.Go(func() error {
				return serv.Setup(ctx)
			})
		}
	})
}
