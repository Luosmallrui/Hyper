package cache

import (
	"Hyper/models"
	"Hyper/pkg/jsonutil"
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const lastMessageCacheKey = "im:message:last_message"

type MessageStorage struct {
	redis *redis.Client
}

func NewMessageStorage(rds *redis.Client) *MessageStorage {
	return &MessageStorage{rds}
}

func (m *MessageStorage) Set(ctx context.Context, talkType int, sender int, receive int, message *LastCacheMessage) error {
	text := jsonutil.Encode(message)

	return m.redis.HSet(ctx, lastMessageCacheKey, m.name(talkType, sender, receive), text).Err()
}

type LastCacheMessage struct {
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

func (m *MessageStorage) BatchGet(ctx context.Context, userId uint64, convs []models.Session) map[uint64]*LastCacheMessage {
	resMap := make(map[uint64]*LastCacheMessage)
	pipe := m.redis.Pipeline()

	for _, c := range convs {
		// 【关键】必须使用和 Set 时完全一样的 name 函数构造 Field
		// 这里的 userId 是当前登录用户，c.PeerId 是对方
		field := m.name(c.SessionType, int(userId), int(c.PeerId))
		pipe.HGet(ctx, lastMessageCacheKey, field)
	}

	cmds, _ := pipe.Exec(ctx)
	for i, cmd := range cmds {
		val, err := cmd.(*redis.StringCmd).Result()
		if err == nil {
			msg := &LastCacheMessage{}
			if json.Unmarshal([]byte(val), msg) == nil {
				// 返回给前端时，用 PeerId 作为 Key，方便 DTO 匹配
				resMap[convs[i].PeerId] = msg
			}
		}
	}
	return resMap
}
func (m *MessageStorage) Get(ctx context.Context, talkType int, sender int, receive int) (*LastCacheMessage, error) {

	res, err := m.redis.HGet(ctx, lastMessageCacheKey, m.name(talkType, sender, receive)).Result()
	if err != nil {
		return nil, err
	}

	msg := &LastCacheMessage{}
	if err = jsonutil.Decode(res, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func (m *MessageStorage) MGet(ctx context.Context, fields []string) ([]*LastCacheMessage, error) {

	res := m.redis.HMGet(ctx, lastMessageCacheKey, fields...)

	items := make([]*LastCacheMessage, 0)
	for _, item := range res.Val() {
		if val, ok := item.(string); ok {
			msg := &LastCacheMessage{}
			if err := jsonutil.Decode(val, msg); err != nil {
				return nil, err
			}

			items = append(items, msg)
		}
	}

	return items, nil
}

func (m *MessageStorage) name(talkType int, sender int, receive int) string {
	if talkType == 2 {
		sender = 0
	}

	if sender > receive {
		sender, receive = receive, sender
	}

	return fmt.Sprintf("%d_%d_%d", talkType, sender, receive)
}
