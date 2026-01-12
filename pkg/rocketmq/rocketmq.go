package rocketmq

import (
	"Hyper/config"
	"Hyper/pkg/log"
	"Hyper/types"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/apache/rocketmq-client-go/v2/rlog"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"

	rmq_client "github.com/apache/rocketmq-clients/golang/v5"
	"go.uber.org/zap"
)

var wg sync.WaitGroup

const (
	Topic         = "IM_CHAT_MSGS"
	ConsumerGroup = "IM_STORAGE_GROUP"
	// Endpoint 填写腾讯云提供的接入地址
	Endpoint = "rmq-163wpoz74o.rocketmq.cd.public.tencenttdmq.com:8080"
	// AccessKey 添加配置的ak
	AccessKey = "ak163wpoz74o167d5c7fb947"
	// SecretKey 添加配置的sk
	SecretKey = "skcbcfbab1a66254c3"
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

var defaultLoggerOnce sync.Once

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
	p, err := rmq_client.NewProducer(&rmq_client.Config{Endpoint: "rmq-163wpoz74o.rocketmq.cd.public.tencenttdmq.com:8080",
		Credentials: &credentials.SessionCredentials{AccessKey: AccessKey, AccessSecret: SecretKey}},
		rmq_client.WithTopics(types.ImTopicChat))
	if err != nil {
		log.L.Fatal("Failed to create producer", zap.Error(err))
	}
	err = p.Start()
	if err != nil {
		log.L.Fatal("Failed to start producer", zap.Error(err))
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
	c, err := rmq_client.NewSimpleConsumer(&rmq_client.Config{
		Endpoint:      Endpoint,
		ConsumerGroup: ConsumerGroup,
		Credentials: &credentials.SessionCredentials{
			AccessKey:    AccessKey,
			AccessSecret: SecretKey,
		},
	},
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
