package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type CaptchaStorage struct {
	redis *redis.Client
}

func NewCaptchaStorage(redis *redis.Client) *CaptchaStorage {
	return &CaptchaStorage{redis: redis}
}

func (c *CaptchaStorage) Set(ctx context.Context, id string, value string) error {
	return c.redis.SetEx(ctx, c.name(id), value, 3*time.Minute).Err()
}

func (c *CaptchaStorage) Get(ctx context.Context, id string, clear bool) string {
	// clear 时使用 GETDEL 原子地完成读取并删除
	if clear {
		return c.redis.GetDel(ctx, c.name(id)).Val()
	}

	return c.redis.Get(ctx, c.name(id)).Val()
}

func (c *CaptchaStorage) Verify(ctx context.Context, id, answer string, clear bool) bool {
	return c.Get(ctx, id, clear) == answer
}

func (c *CaptchaStorage) name(id string) string {
	return fmt.Sprintf("im:auth:captcha:%s", id)
}
