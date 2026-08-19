package process

import (
	"Hyper/pkg/log"
	"Hyper/types"
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
	HealthSubscribe       *HealthSubscribe  // 注册健康上报
	MessageSubscribe      *MessageSubscribe /// 注册消息订阅
	NoticeSubscribe       *NoticeSubscribe
	NoteClassifySubscribe *NoteClassifySubscribe
	NoteImageTagSubscribe *NoteImageTagSubscribe
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
// 初始化或启动消费者失败时返回 error，由调用方决定进程去留（不要在这里 log.Fatal 杀死进程）
func (c *Server) Start(eg *errgroup.Group, ctx context.Context) error {
	var startErr error
	once.Do(func() {
		for _, process := range c.items {
			if err := process.Init(); err != nil {
				startErr = fmt.Errorf("注册 Topic 失败: %w", err)
				return
			}
		}

		for _, process := range c.items {
			serv := process
			eg.Go(func() error {
				return serv.Setup(ctx)
			})
		}

		if err := c.MqConsumer.Start(); err != nil {
			startErr = fmt.Errorf("failed to start consumer: %w", err)
			return
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
						mvs, err := c.MqConsumer.Receive(ctx, maxMessageNum, invisibleDuration)
						if err != nil {
							if ctx.Err() != nil {
								return nil // 正常退出
							}
							// Receive 失败不能热自旋，记日志并退避后重试
							log.L.Error("receive message error", zap.Error(err))
							select {
							case <-ctx.Done():
								return nil
							case <-time.After(time.Second):
							}
							continue
						}
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
	return startErr
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
			err = c.NoticeSubscribe.handleSystem(ctx, mv)
		}
	case types.NoteClassifyTopic:
		if c.NoteClassifySubscribe != nil {
			err = c.NoteClassifySubscribe.handleNoteClassify(ctx, mv)
		}
	case types.NoteImageTagTopic:
		if c.NoteImageTagSubscribe != nil {
			err = c.NoteImageTagSubscribe.handleNoteImageTag(ctx, mv)
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
