package rocketmq

import (
	"Hyper/config"
	"Hyper/pkg/log"
	"Hyper/types"
	"os"
	"path/filepath"
	"time"

	"github.com/apache/rocketmq-clients/golang/v5/credentials"

	rmq_client "github.com/apache/rocketmq-clients/golang/v5"
	"go.uber.org/zap"
)

var (
	// maximum waiting time for receive func
	awaitDuration = time.Second * 5
	// maximum number of messages received at one time
	maxMessageNum int32 = 16
	// invisibleDuration should > 20s
	invisibleDuration = time.Second * 20
	// receive concurrency
	receiveConcurrency = 6
)

func init() {

}

func InitProducer(cfg *config.RocketMQConfig) rmq_client.Producer {
	if len(cfg.NameServer) == 0 {
		log.L.Fatal("rocketmq nameserver is empty")
	}
	dir, _ := os.Getwd()
	logPath := filepath.Join(dir, "logs")

	_ = os.MkdirAll(logPath, 0755)

	os.Setenv("rmq.client.logRoot", logPath)
	os.Setenv("mq.consoleAppender.enabled", "true")
	os.Setenv("rmq.client.logRoot", logPath)
	os.Setenv("rocketmq.client.logRoot", logPath)
	rmq_client.EnableSsl = false
	rmq_client.ResetLogger()
	rmqConfig := &rmq_client.Config{
		Endpoint: cfg.NameServer[0],
		Credentials: &credentials.SessionCredentials{
			AccessKey:    cfg.Ak,
			AccessSecret: cfg.Sk,
		},
	}
	p, err := rmq_client.NewProducer(rmqConfig, rmq_client.WithTopics(
		types.ImTopicChat,
		types.NoteClassifyTopic,
		types.NoteImageTagTopic,
	))
	if err != nil {
		log.L.Fatal("Failed to create producer", zap.Error(err))
	}
	if err := p.Start(); err != nil {
		log.L.Fatal("Failed to start producer", zap.Error(err))
	}
	return p
}

func InitConsumer(cfg *config.RocketMQConfig) rmq_client.SimpleConsumer {
	//os.Setenv("mq.consoleAppender.enabled", "true")
	if len(cfg.NameServer) == 0 {
		log.L.Fatal("rocketmq nameserver is empty")
	}
	dir, _ := os.Getwd()
	logPath := filepath.Join(dir, "logs") // 结果类似 /Users/name/project/logs

	// 确保在设置变量前，手动创建好这个目录
	_ = os.MkdirAll(logPath, 0755)

	// 必须在 ResetLogger 之前设置
	os.Setenv("rmq.client.logRoot", logPath)
	os.Setenv("mq.consoleAppender.enabled", "true")
	os.Setenv("rmq.client.logRoot", logPath)
	os.Setenv("rocketmq.client.logRoot", logPath)
	rmq_client.EnableSsl = false
	rmq_client.ResetLogger()
	rmqConfig := &rmq_client.Config{Endpoint: cfg.NameServer[0], ConsumerGroup: cfg.Consumer.Group}
	if cfg.Ak != "" && cfg.Sk != "" {
		rmqConfig.Credentials = &credentials.SessionCredentials{AccessKey: cfg.Ak, AccessSecret: cfg.Sk}
	}
	c, err := rmq_client.NewSimpleConsumer(rmqConfig,
		rmq_client.WithSimpleAwaitDuration(awaitDuration),
		rmq_client.WithSimpleSubscriptionExpressions(map[string]*rmq_client.FilterExpression{
			"IM_CHAT_MSGS":          rmq_client.SUB_ALL,
			"HYPER_SYSTEM_MSGS":     rmq_client.SUB_ALL,
			types.NoteClassifyTopic: rmq_client.SUB_ALL,
			types.NoteImageTagTopic: rmq_client.SUB_ALL,
		}),
	)
	if err != nil {
		log.L.Fatal("Failed to create consumer", zap.Error(err))
	}
	return c
}
