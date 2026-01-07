package cache

import (
	"Hyper/models"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// 未读消息过期时间 - 14天
const unreadExpireAt = 14 * 24 * time.Hour

type UnreadStorage struct {
	redis *redis.Client
}

func NewUnreadStorage(rds *redis.Client) *UnreadStorage {
	return &UnreadStorage{rds}
}

// Incr 消息未读数自增
// @params uid     用户ID
// @params mode    对话模式 1私信 2群聊
// @params sender  发送者ID(群ID)
func (u *UnreadStorage) Incr(ctx context.Context, uid, mode, sender int) {
	pipe := u.redis.Pipeline()
	u.PipeIncr(ctx, pipe, uid, mode, sender)
	_, _ = pipe.Exec(ctx)
}

// PipeIncr 消息未读数自增
// @params uid     用户ID
// @params mode    对话模式 1私信 2群聊
// @params sender  发送者ID(群ID)
func (u *UnreadStorage) PipeIncr(ctx context.Context, pipe redis.Pipeliner, uid, mode, sender int) {
	name := u.name(uid, mode, sender)
	pipe.Incr(ctx, name)
	pipe.Expire(ctx, name, unreadExpireAt)
}

// Get 获取消息未读数
// @params uid     用户ID
// @params mode    对话模式 1私信 2群聊
// @params sender  发送者ID(群ID)
func (u *UnreadStorage) Get(ctx context.Context, uid, mode, sender int) int {
	i, err := u.redis.Get(ctx, u.name(uid, mode, sender)).Int()
	if err != nil {
		return 0
	}

	return i
}

// Del 删除消息未读数
// @params uid     用户ID
// @params mode    对话模式 1私信 2群聊
// @params sender  发送者ID(群ID)
func (u *UnreadStorage) Del(ctx context.Context, uid, mode, sender int) {
	u.redis.Del(ctx, u.name(uid, mode, sender))
}

// Reset 消息未读数重置
// @params uid     用户ID
// @params mode    对话模式 1私信 2群聊
// @params sender  发送者ID(群ID)
func (u *UnreadStorage) Reset(ctx context.Context, uid, mode, sender int) {
	u.Del(ctx, uid, mode, sender)
}

// 未读数缓存
// mode, uid, sender int
// im:unread:uid:mode_sender
func (u *UnreadStorage) name(uid, mode, sender int) string {
	return fmt.Sprintf("im:unread:%d:%d_%d", uid, mode, sender)
}

func (u *UnreadStorage) BatchGet(ctx context.Context, userId uint64, convs []models.Session) map[uint64]uint32 {
	resMap := make(map[uint64]uint32)
	pipe := u.redis.Pipeline()

	// 假设未读数 Key 是 "unread:123"
	key := fmt.Sprintf("unread:%d", userId)
	for _, c := range convs {
		// 【关键】这里的 field 构造逻辑必须和 PipeIncr 时一致
		// 如果 PipeIncr 存的是 "1_1001" (talkType_peerId)
		field := fmt.Sprintf("%d_%d", c.SessionType, c.PeerId)
		pipe.HGet(ctx, key, field)
	}

	cmds, _ := pipe.Exec(ctx)
	for i, cmd := range cmds {
		val, err := cmd.(*redis.StringCmd).Int()
		if err == nil {
			resMap[convs[i].PeerId] = uint32(val)
		}
	}
	return resMap
}
