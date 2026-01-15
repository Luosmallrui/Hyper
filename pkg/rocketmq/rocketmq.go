package rocketmq

import (
	"Hyper/config"
	"Hyper/pkg/log"
	"Hyper/types"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/apache/rocketmq-client-go/v2/rlog"
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
	rlog.SetLogLevel("error")
	rlog.SetOutputPath("./logss/a.log")
}

func InitProducer(cfg *config.RocketMQConfig) rmq_client.Producer {
	//os.Setenv("mq.consoleAppender.enabled", "true")
	dir, _ := os.Getwd()
	logPath := filepath.Join(dir, "logs") // 结果类似 /Users/name/project/logs

	// 确保在设置变量前，手动创建好这个目录
	_ = os.MkdirAll(logPath, 0755)

	fmt.Println("log path:", logPath)
	// 必须在 ResetLogger 之前设置
	os.Setenv("rmq.client.logRoot", logPath)
	os.Setenv("mq.consoleAppender.enabled", "true")
	os.Setenv("rmq.client.logRoot", logPath)
	os.Setenv("rocketmq.client.logRoot", logPath)
	rmq_client.ResetLogger()
	rmqConfig := &rmq_client.Config{Endpoint: cfg.NameServer[0]}
	if cfg.Ak != "" && cfg.Sk != "" {
		rmqConfig.Credentials = &credentials.SessionCredentials{AccessKey: cfg.Ak, AccessSecret: cfg.Sk}
	}
	p, err := rmq_client.NewProducer(rmqConfig, rmq_client.WithTopics(types.ImTopicChat))
	if err != nil {
		log.L.Info("Failed to create producer", zap.Error(err))
	}
	err = p.Start()
	if err != nil {
		log.L.Info("Failed to start producer", zap.Error(err))
	}
	return p
}

func InitConsumer(cfg *config.RocketMQConfig) rmq_client.SimpleConsumer {
	//os.Setenv("mq.consoleAppender.enabled", "true")
	dir, _ := os.Getwd()
	logPath := filepath.Join(dir, "logs") // 结果类似 /Users/name/project/logs

	// 确保在设置变量前，手动创建好这个目录
	_ = os.MkdirAll(logPath, 0755)

	// 必须在 ResetLogger 之前设置
	os.Setenv("rmq.client.logRoot", logPath)
	os.Setenv("mq.consoleAppender.enabled", "true")
	os.Setenv("rmq.client.logRoot", logPath)
	os.Setenv("rocketmq.client.logRoot", logPath)
	rmq_client.ResetLogger()
	rmqConfig := &rmq_client.Config{Endpoint: cfg.NameServer[0], ConsumerGroup: cfg.Consumer.Group}
	if cfg.Ak != "" && cfg.Sk != "" {
		rmqConfig.Credentials = &credentials.SessionCredentials{AccessKey: cfg.Ak, AccessSecret: cfg.Sk}
	}
	c, err := rmq_client.NewSimpleConsumer(rmqConfig,
		rmq_client.WithSimpleAwaitDuration(awaitDuration),
		rmq_client.WithSimpleSubscriptionExpressions(map[string]*rmq_client.FilterExpression{
			"IM_CHAT_MSGS":      rmq_client.SUB_ALL,
			"HYPER_SYSTEM_MSGS": rmq_client.SUB_ALL,
		}),
	)
	if err != nil {
		log.L.Fatal("Failed to create consumer", zap.Error(err))
	}
	return c
}
