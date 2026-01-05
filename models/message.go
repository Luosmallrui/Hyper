package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Message struct {
	MsgID       string `json:"msg_id"`        // 消息唯一ID (Snowflake)
	SenderID    string `json:"sender_id"`     // 发送者UID
	TargetID    string `json:"target_id"`     // 接收者UID或群ID
	SessionType int    `json:"session_type"`  // 1-单聊,2-群聊,3-系统通知
	MsgType     int    `json:"msg_type"`      // 消息类型
	Content     string `json:"content"`       // 消息正文
	ParentMsgID string `json:"parent_msg_id"` // 回复消息ID
	Timestamp   int64  `json:"timestamp"`     // 发送时间戳(毫秒)
	Status      int    `json:"status"`        // 状态
	Ext         string `json:"ext,omitempty"` // JSON 字符串 // 扩展字段
}

// TableName 指定表名
func (Message) TableName() string {
	return "im_message"
}

// MessageExt 扩展字段，支持 JSON 存储
type MessageExt map[string]interface{}

// 实现 gorm 的接口，支持 json 类型读写
func (m MessageExt) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *MessageExt) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, m)
}

// ImSingleMessage 单聊消息结构体
type ImSingleMessage struct {
	Id          int64      `gorm:"primaryKey;column:id" json:"msg_id,string"`
	ClientMsgId string     `gorm:"uniqueIndex:uk_client_msg_id;column:client_msg_id" json:"client_msg_id"`
	SessionHash int64      `gorm:"index:idx_session_time;column:session_hash" json:"session_hash"` // 内部索引，不返回前端
	SessionId   string     `gorm:"column:session_id" json:"session_id"`
	SenderId    int64      `gorm:"column:sender_id" json:"sender_id,string"`
	TargetId    int64      `gorm:"column:target_id" json:"target_id,string"`
	MsgType     int        `gorm:"column:msg_type;default:1" json:"msg_type"` //1-文本，2-图片，3-视频，4-语音，5-文件等）
	Content     string     `gorm:"column:content" json:"content"`             //消息的正文。
	ParentMsgId int64      `gorm:"column:parent_msg_id;default:0" json:"parent_msg_id,string"`
	Status      int        `gorm:"column:status;default:1" json:"status"` //（1-正常，2-已撤回，3-逻辑删除）。
	Ext         MessageExt `gorm:"type:json;column:ext" json:"ext,omitempty"`
	CreatedAt   int64      `gorm:"index:idx_session_time;column:created_at" json:"timestamp"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"-"`
}

// TableName 指定数据库表名
func (ImSingleMessage) TableName() string {
	return "im_single_messages"
}
