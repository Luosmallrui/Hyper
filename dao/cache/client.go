package cache

import (
	"Hyper/config"
	"context"
	"fmt"
	"strconv"
	"time"

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
	return c.Set(ctx, sid, uid, clientId)
}

func (c *ClientStorage) UnBind(ctx context.Context, sid, channel string, clientId int64, uid int) error {
	return c.Del(ctx, sid, uid, clientId)
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
	uidInt, err := strconv.Atoi(uid)
	if err != nil {
		return false
	}
	val, err := c.redis.SCard(ctx, c.userKey(sid, uidInt)).Result()
	return err == nil && val > 0
}

// GetUidFromClientIds 获取当前节点用户ID关联的客户端ID
// @params sid      服务ID
// @params channel  渠道分组
// @params uid      用户ID
func (c *ClientStorage) GetUidFromClientIds(ctx context.Context, sid, channel, uid string) []int64 {
	cids := make([]int64, 0)

	uidInt, err := strconv.Atoi(uid)
	if err != nil {
		return cids
	}
	items, err := c.redis.SMembers(ctx, c.userKey(sid, uidInt)).Result()
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
	uid, err := c.redis.HGet(ctx, c.clientKey(sid), fmt.Sprintf("%d", clientId)).Result()
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(uid, 10, 64)
}

func (c *ClientStorage) GetUserRoute(ctx context.Context, uid int) (map[string][]string, error) {
	// 1. 一键获取该用户所有的 clientId 和对应的 serverId
	// Key: im:user:location:100 -> { "c1": "sid_A", "c2": "sid_A", "c3": "sid_B" }
	results, err := c.redis.HGetAll(ctx, c.userLocationKey(uid)).Result()
	if err != nil {
		return nil, err
	}

	// 2. 在内存中按 sid 进行分组
	routeMap := make(map[string][]string)
	for cid, sid := range results {
		routeMap[sid] = append(routeMap[sid], cid)
	}

	// 返回结果示例: map["sid_A"]: ["c1", "c2"], map["sid_B"]: ["c3"]
	return routeMap, nil
}
func (c *ClientStorage) Set(ctx context.Context, sid string, uid int, clientId int64) error {
	// 转换 clientId 为 string 方便 Redis 存储
	cidStr := strconv.FormatInt(clientId, 10)

	_, err := c.redis.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		// 1. 全局位置索引 (Hash)
		// Key: im:user:location:100, Field: 123456789, Value: ws-01
		// 作用：消息路由时通过 uid 快速定位 sid
		pipe.HSet(ctx, c.userLocationKey(uid), cidStr, sid)

		//  节点内的用户连接详情 (Set)
		// Key: im:server:ws-01:clients:100, Value: [123456789, ...]
		// 作用：消息到达 ws-01 后，找到该 uid 对应的所有本地连接
		pipe.SAdd(ctx, c.userKey(sid, uid), cidStr)

		// 节点内客户端ID -> 用户ID 映射 (Hash)
		// Key: im:server:ws-01:client, Field: 123456789, Value: 100
		pipe.HSet(ctx, c.clientKey(sid), cidStr, uid)

		// Key 设置过期时间，配合心跳续期，防止僵尸数据（异常断连时由 TTL 兜底清理）
		pipe.Expire(ctx, c.userLocationKey(uid), 24*time.Hour)
		pipe.Expire(ctx, c.userKey(sid, uid), 24*time.Hour)
		pipe.Expire(ctx, c.clientKey(sid), 24*time.Hour)
		return nil
	})
	return err
}

func (c *ClientStorage) userLocationKey(uid int) string {
	return fmt.Sprintf("im:user:location:%d", uid)
}
func (c *ClientStorage) clientKey(sid string) string {
	return fmt.Sprintf("im:server:%s:client", sid)
}

func (c *ClientStorage) userKey(sid string, uid int) string {
	return fmt.Sprintf("im:server:%s:clients:%d", sid, uid)
}

func (c *ClientStorage) Del(ctx context.Context, sid string, uid int, clientId int64) error {
	cidStr := strconv.FormatInt(clientId, 10)

	_, err := c.redis.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		// 1. 移除全局位置中的特定设备
		pipe.HDel(ctx, c.userLocationKey(uid), cidStr)

		// 2. 移除节点详情中的特定设备
		pipe.SRem(ctx, c.userKey(sid, uid), cidStr)

		// 3. 移除客户端ID -> 用户ID 映射
		pipe.HDel(ctx, c.clientKey(sid), cidStr)

		return nil
	})
	return err
}
