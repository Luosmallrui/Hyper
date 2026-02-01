package dao

import (
	"context"
	"errors"
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

// DeleteSession 删除某个用户的某个会话
func (d *SessionDAO) DeleteSession(ctx context.Context, userID uint64, sessionType int, peerID uint64) error {
	return d.db.WithContext(ctx).
		Where("user_id = ? AND session_type = ? AND peer_id = ?", userID, sessionType, peerID).
		Delete(&models.Session{}).Error
}

// DeleteSessionsByPeer 删除某个群/对端的所有会话（解散群用）
func (d *SessionDAO) DeleteSessionsByPeer(ctx context.Context, sessionType int, peerID uint64) error {
	return d.db.WithContext(ctx).
		Where("session_type = ? AND peer_id = ?", sessionType, peerID).
		Delete(&models.Session{}).Error
}

func (d *SessionDAO) GetUnreadNum(ctx context.Context, userID int) (int64, error) {
	var total int64

	err := d.db.WithContext(ctx).
		Model(&models.Session{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(unread_count), 0)").
		Scan(&total).Error

	if err != nil {
		return 0, err
	}

	return total, nil
}
func (d *SessionDAO) WithDB(db *gorm.DB) *SessionDAO {
	nd := *d
	nd.db = db
	return &nd
}

// EnsureSession: 确保某个用户对某个 peer 的会话存在（幂等）
// sessionType: 1=单聊(peerId=对方user_id), 2=群聊(peerId=group_id)
func (d *SessionDAO) EnsureSession(ctx context.Context, tx *gorm.DB, uid uint64, sessionType int, peerId uint64) (uint64, error) {
	if tx == nil {
		tx = d.db
	}
	if tx == nil {
		return 0, errors.New("SessionDAO db 未初始化")
	}

	now := time.Now()
	nowMs := now.UnixMilli()

	s := &models.Session{
		UserId:      uid,
		SessionType: sessionType,
		PeerId:      peerId,

		LastMsgId:      0,
		LastMsgType:    0,
		LastMsgContent: "",
		LastMsgTime:    nowMs,

		UnreadCount: 0,
		IsTop:       0,
		IsMute:      0,

		CreatedAt: now,
		UpdatedAt: now,
	}

	// 幂等插入：已存在就不插
	if err := tx.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "user_id"},
				{Name: "session_type"},
				{Name: "peer_id"},
			},
			DoNothing: true,
		}).
		Create(s).Error; err != nil {
		return 0, err
	}

	// 如果是新插入，s.Id 会回填；如果 DoNothing，s.Id 可能还是 0，需要查一次
	if s.Id != 0 {
		return s.Id, nil
	}

	var existing models.Session
	if err := tx.WithContext(ctx).
		Where("user_id = ? AND session_type = ? AND peer_id = ?", uid, sessionType, peerId).
		First(&existing).Error; err != nil {
		return 0, err
	}
	return existing.Id, nil
}
