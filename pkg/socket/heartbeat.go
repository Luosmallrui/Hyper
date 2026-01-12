package socket

import (
	"context"
	"errors"
	"strconv"
	"time"

	"Hyper/pkg/timewheel"
)

const (
	heartbeatInterval = 10 // 心跳检测间隔时间
	heartbeatTimeout  = 35 // 心跳检测超时时间（超时时间是隔间检测时间的2.5倍以上）
)

var health *heartbeat

// 客户端心跳管理
type heartbeat struct {
	timeWheel *timewheel.SimpleTimeWheel[*Client]
}

func init() {
	health = &heartbeat{}
	health.timeWheel = timewheel.NewSimpleTimeWheel[*Client](1*time.Second, 100, health.handle)
}

func (h *heartbeat) Start(ctx context.Context) error {

	go h.timeWheel.Start()

	<-ctx.Done()

	h.timeWheel.Stop()

	return errors.New("heartbeat exit")
}

func (h *heartbeat) insert(c *Client) {
	h.timeWheel.Add(strconv.FormatInt(c.cid, 10), c, time.Duration(heartbeatInterval)*time.Second)
}

func (h *heartbeat) delete(c *Client) {
	h.timeWheel.Remove(strconv.FormatInt(c.cid, 10))
}

func (h *heartbeat) handle(timeWheel *timewheel.SimpleTimeWheel[*Client], key string, c *Client) {
	if c.Closed() {
		return
	}

	now := time.Now().Unix()
	interval := int(now - c.lastTime)

	if interval > heartbeatTimeout {
		c.Close(2000, "心跳检测超时")
		return
	}

	if interval >= heartbeatInterval {
		_ = c.Write(&ClientResponse{Event: "ping"})
	}

	nextCheck := heartbeatInterval
	if interval < heartbeatInterval {
		nextCheck = heartbeatInterval - interval + 1
	} else {
		nextCheck = heartbeatInterval
	}
	if nextCheck <= 0 {
		nextCheck = 1
	}

	timeWheel.Add(key, c, time.Duration(nextCheck)*time.Second)
}
