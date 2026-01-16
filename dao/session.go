package dao

import (
	"context"

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
