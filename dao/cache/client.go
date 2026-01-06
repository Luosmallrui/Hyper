package cache

import (
	"Hyper/config"
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type ClientStorage struct {
	redis   *redis.Client
	config  *config.Config
	storage *ServerStorage
}

func NewClientStorage(redis *redis.Client, config *config.Config, storage *ServerStorage) *ClientStorage {
	return &ClientStorage{redis: redis, config: config, storage: storage}
}

func (c *ClientStorage) Bind(ctx context.Context, sid, channel string, clientId int64, uid int) error {
	return c.set(ctx, sid, channel, clientId, uid)
}

func (c *ClientStorage) UnBind(ctx context.Context, sid, channel string, clientId int64) error {
	return c.del(ctx, sid, channel, strconv.FormatInt(clientId, 10))
}

// IsOnline 判断客户端是否在线[所有部署机器]
// @params channel  渠道分组
// @params uid      用户ID
func (c *ClientStorage) IsOnline(ctx context.Context, channel, uid string) bool {
	for _, sid := range c.storage.All(ctx, 1) {
		if c.IsCurrentServerOnline(ctx, sid, channel, uid) {
			return true
		}
	}

	return false
}

// IsCurrentServerOnline 判断当前节点是否在线
// @params sid      服务ID
// @params channel  渠道分组
// @params uid      用户ID
func (c *ClientStorage) IsCurrentServerOnline(ctx context.Context, sid, channel, uid string) bool {
	val, err := c.redis.SCard(ctx, c.userKey(sid, channel, uid)).Result()
	return err == nil && val > 0
}

// GetUidFromClientIds 获取当前节点用户ID关联的客户端ID
// @params sid      服务ID
// @params channel  渠道分组
// @params uid      用户ID
func (c *ClientStorage) GetUidFromClientIds(ctx context.Context, sid, channel, uid string) []int64 {
	cids := make([]int64, 0)

	items, err := c.redis.SMembers(ctx, c.userKey(sid, channel, uid)).Result()
	if err != nil {
		return cids
	}

	for _, cid := range items {
		if cid, err := strconv.ParseInt(cid, 10, 64); err == nil {
			cids = append(cids, cid)
		}
	}

	return cids
}

// GetClientIdFromUid 获取客户端ID关联的用户ID
// @params sid     服务节点ID
// @params channel 渠道分组
// @params cid     客户端ID
func (c *ClientStorage) GetClientIdFromUid(ctx context.Context, sid, channel string, clientId int64) (int64, error) {
	uid, err := c.redis.HGet(ctx, c.clientKey(sid, channel), fmt.Sprintf("%d", clientId)).Result()
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(uid, 10, 64)
}

// 设置客户端与用户绑定关系
// @params channel  渠道分组
// @params fd       客户端连接ID
// @params id       用户ID
func (c *ClientStorage) set(ctx context.Context, sid, channel string, clientId int64, uid int) error {
	_, err := c.redis.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		// 命令1: 存储客户端与用户的映射
		pipe.HSet(ctx, c.clientKey(sid, channel), clientId, uid)
		// 命令2: 存储用户的所有客户端
		pipe.SAdd(ctx, c.userKey(sid, channel, strconv.Itoa(uid)), clientId)
		pipe.HIncrBy(ctx, c.userLocationKey(uid), sid, 1)
		return nil
	})
	return err
}

func (c *ClientStorage) userLocationKey(uid int) string {
	return fmt.Sprintf("ws:user:location:%d", uid)
}
func (c *ClientStorage) clientKey(sid, channel string) string {
	return fmt.Sprintf("ws:%s:%s:client", sid, channel)
}

func (c *ClientStorage) userKey(sid, channel, uid string) string {
	return fmt.Sprintf("ws:%s:%s:user:%s", sid, channel, uid)
}

// 删除客户端与用户绑定关系
// @params channel  渠道分组
// @params fd       客户端连接ID
func (c *ClientStorage) del(ctx context.Context, sid, channel, fd string) error {
	key := c.clientKey(sid, channel)

	// 1. 先获取该客户端对应的 UID (string 类型)
	uidStr, err := c.redis.HGet(ctx, key, fd).Result()
	if err != nil {
		return err
	}

	// 2. 将 string 转换为 int，以便调用 userLocationKey
	uid, err := strconv.Atoi(uidStr)
	if err != nil {
		log.Printf("[ERROR] uid 转换失败: %v", err)
		return err
	}

	_, err = c.redis.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		// 3. 删除客户端与用户的映射
		pipe.HDel(ctx, key, fd)

		// 4. 从用户设备集合中移除 (注意这里传 uidStr 或 uid 视你 userKey 定义而定)
		pipe.SRem(ctx, c.userKey(sid, channel, uidStr), fd)

		// 5. 核心：用户位置计数减 1
		locationKey := c.userLocationKey(uid)
		pipe.HIncrBy(ctx, locationKey, sid, -1)

		return nil
	})
	// 6. 延时检查：如果该 sid 计数归零，彻底删除该 field
	// 放在 Pipeline 外执行是为了确保获取到最新的计数
	count, _ := c.redis.HGet(ctx, c.userLocationKey(uid), sid).Int64()
	if count <= 0 {
		c.redis.HDel(ctx, c.userLocationKey(uid), sid)
	}
	return err
}
