package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisLock struct {
	redis *redis.Client
}

func NewRedisLock(rds *redis.Client) *RedisLock {
	return &RedisLock{rds}
}

// Lock 获取 redis 锁，成功时返回持有 token，解锁时必须携带
func (r *RedisLock) Lock(ctx context.Context, name string, expire int) (string, bool) {
	token := uuid.NewString()
	if !r.redis.SetNX(ctx, r.name(name), token, time.Duration(expire)*time.Second).Val() {
		return "", false
	}
	return token, true
}

// UnLock 释放 redis 锁，仅当 token 与持锁者一致时才删除
func (r *RedisLock) UnLock(ctx context.Context, name, token string) bool {
	script := `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("DEL", KEYS[1])
	else
		return 0
	end`

	n, err := r.redis.Eval(ctx, script, []string{r.name(name)}, token).Int()
	return err == nil && n == 1
}

func (r *RedisLock) name(name string) string {
	return fmt.Sprintf("im:lock:%s", name)
}
