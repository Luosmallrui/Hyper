package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	rmq "github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
)

func main() {
	dir, _ := os.Getwd()
	logPath := filepath.Join(dir, "logs")
	_ = os.MkdirAll(logPath, 0755)

	os.Setenv("rmq.client.logRoot", logPath)
	os.Setenv("rocketmq.client.logRoot", logPath)
	os.Setenv("mq.consoleAppender.enabled", "true")
	rmq.EnableSsl = false
	rmq.ResetLogger()

	endpoint := envDefault("RMQ_ENDPOINT", "8.156.86.220:19081")
	topic := envDefault("RMQ_TOPIC", "IM_CHAT_MSGS")
	ak := os.Getenv("RMQ_AK")
	sk := os.Getenv("RMQ_SK")
	fmt.Println("connecting to:", endpoint)
	fmt.Println("topic:", topic)

	cfg := &rmq.Config{
		Endpoint: endpoint,
		Credentials: &credentials.SessionCredentials{
			AccessKey:    ak,
			AccessSecret: sk,
		},
	}
	if ak != "" || sk != "" {
		if ak == "" || sk == "" {
			panic("RMQ_AK and RMQ_SK must be set together")
		}
	}

	p, err := rmq.NewProducer(cfg, rmq.WithTopics(topic))
	if err != nil {
		panic(fmt.Errorf("new producer: %w", err))
	}

	if err := p.Start(); err != nil {
		panic(fmt.Errorf("start producer: %w", err))
	}
	defer p.GracefulStop()

	msg := &rmq.Message{
		Topic: topic,
		Body:  buildBody(topic),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := p.Send(ctx, msg)
	if err != nil {
		panic(fmt.Errorf("send message: %w", err))
	}

	fmt.Println("send success:", resp)
	time.Sleep(time.Second)
}

func envDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func buildBody(topic string) []byte {
	if topic == "HYPER_SYSTEM_MSGS" {
		body, _ := json.Marshal(map[string]any{
			"type": "platform_message",
			"data": map[string]any{
				"message_id": 0,
				"target_id":  0,
				"title":      "RocketMQ 测试",
				"content":    "hello rocketmq",
				"type":       "system",
				"target":     "test",
				"created_at": time.Now().Format(time.RFC3339),
			},
		})
		return body
	}
	body, _ := json.Marshal(map[string]any{
		"msg_id":       fmt.Sprintf("%d", time.Now().UnixNano()),
		"sender_id":    "0",
		"target_id":    "0",
		"session_type": 1,
		"session_hash": 0,
		"session_id":   "test",
		"msg_type":     1,
		"content":      "hello rocketmq",
		"timestamp":    time.Now().UnixMilli(),
		"status":       1,
		"ext":          map[string]any{"source": "test_rmq"},
		"channel":      "test",
	})
	return body
}
