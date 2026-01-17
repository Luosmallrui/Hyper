package dao

import (
	"context"
	"time"

	"Hyper/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SessionDAO struct {
	db *gorm.DB
}

func NewSessionDAO(db *gorm.DB) *SessionDAO {
	return &SessionDAO{db: db}
}

// BatchUpsert: 批量 upsert 到 im_session
// 依赖唯一索引 uk_user_session(user_id, session_type, peer_id)
func (d *SessionDAO) BatchUpsert(ctx context.Context, rows []models.Session) error {
	if len(rows) == 0 {
		return nil
	}

	return d.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "user_id"},
			{Name: "session_type"},
			{Name: "peer_id"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"last_msg_id":      gorm.Expr("VALUES(last_msg_id)"),
			"last_msg_type":    gorm.Expr("VALUES(last_msg_type)"),
			"last_msg_content": gorm.Expr("VALUES(last_msg_content)"),
			"last_msg_time":    gorm.Expr("VALUES(last_msg_time)"),
			// 未读：累加（发送者传 0，其他成员传 1）
			"unread_count": gorm.Expr("unread_count + VALUES(unread_count)"),
			"updated_at":   gorm.Expr("VALUES(updated_at)"),
		}),
	}).Create(&rows).Error
}
func (d *SessionDAO) UpsertSessionSettings(
	ctx context.Context,
	userID uint64,
	sessionType int,
	peerID uint64,
	isTop int,
	isMute int,
) error {
	now := time.Now()

	// 如果会话不存在(即刚加上或者刚建群还没开始聊天)，也能插入一条“最小合法记录”
	// im_session 表 last_msg_* 都是 NOT NULL，所以要给默认值
	row := models.Session{
		UserId:      userID,
		SessionType: sessionType,
		PeerId:      peerID,

		LastMsgId:      0,
		LastMsgType:    0,
		LastMsgContent: "",
		LastMsgTime:    0,

		UnreadCount: 0,
		IsTop:       isTop,
		IsMute:      isMute,

		CreatedAt: now,
		UpdatedAt: now,
	}

	return d.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "user_id"},
				{Name: "session_type"},
				{Name: "peer_id"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"is_top":     isTop,
				"is_mute":    isMute,
				"updated_at": now,
			}),
		}).
		Create(&row).Error
}
