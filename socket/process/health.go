package process

import (
	"Hyper/dao/cache"
	"Hyper/pkg/log"
	"Hyper/pkg/server"
	"context"
	"time"
)

type HealthSubscribe struct {
	storage *cache.ServerStorage
}

func NewHealthSubscribe(storage *cache.ServerStorage) *HealthSubscribe {
	return &HealthSubscribe{storage}
}

func (s *HealthSubscribe) Init() error {

	return nil
}

func (s *HealthSubscribe) Setup(ctx context.Context) error {

	log.L.Info("start health subscribe")

	timer := time.NewTicker(5 * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			if err := s.storage.Set(ctx, server.GetServerId(), time.Now().Unix()); err != nil {
				//log.Std().Error(fmt.Sprintf("Websocket HealthSubscribe Report Err: %s", err.Error()))
			}
		}
	}
}
