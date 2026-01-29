package dao

import (
	"Hyper/models"
	"context"
	"time"

	"gorm.io/gorm"
)

type NoteDAO struct {
	Repo[models.Note]
}

func NewNoteDAO(db *gorm.DB) *NoteDAO {
	return &NoteDAO{Repo: NewRepo[models.Note](db)}
}

// Create 创建笔记
func (d *NoteDAO) Create(ctx context.Context, note *models.Note) error {
	return d.Db.WithContext(ctx).Create(note).Error
}

// FindByUserID 根据用户ID查询笔记列表
func (d *NoteDAO) FindByUserID(ctx context.Context, userID uint64, status int, limit, offset int) ([]*models.Note, error) {
	var notes []*models.Note
	err := d.Db.WithContext(ctx).
		// Debug().
		Where("user_id = ? AND status = ?", userID, status).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notes).Error
	return notes, err
}

func (d *NoteDAO) ListNode(ctx context.Context, cursor int64, limit int) (notes []*models.Note, err error) {
	db := d.Db.WithContext(ctx).Model(&models.Note{})

	// 如果前端传了游标（大于0），则查询该时间点之前的数据
	if cursor > 0 {
		// 将纳秒时间戳转回 time.Time 对象
		cursorTime := time.Unix(0, cursor)
		db = db.Where("created_at < ?", cursorTime)
	}
	// 必须按时间倒序排，最新的在前
	err = db.Order("created_at DESC").
		Limit(limit).
		Find(&notes).Error

	return notes, err
}

// UpdateStatus 更新笔记状态
func (d *NoteDAO) UpdateStatus(ctx context.Context, noteID uint64, status int) error {
	return d.Db.WithContext(ctx).
		Model(&models.Note{}).
		Where("id = ?", noteID).
		Update("status", status).Error
}

// FindByIDs 根据 ID 列表查询笔记
func (d *NoteDAO) FindByIDs(ctx context.Context, ids []uint64) ([]*models.Note, error) {
	if len(ids) == 0 {
		return []*models.Note{}, nil
	}
	var notes []*models.Note
	err := d.Db.WithContext(ctx).
		Where("id IN ?", ids).
		Find(&notes).Error
	return notes, err
}

func (d *NoteDAO) GetByID(ctx context.Context, noteID uint64) (*models.Note, error) {
	var note models.Note
	err := d.Db.WithContext(ctx).
		Where("id = ?", noteID).
		First(&note).Error
	return &note, err
}

// dao/note_stats_dao.go

func (d *NoteStatsDAO) IncrementViewCount(ctx context.Context, noteID uint64, delta int) error {
	return d.Db.WithContext(ctx).
		Model(&models.NoteStats{}).
		Where("note_id = ?", noteID).
		UpdateColumn("view_count", gorm.Expr("view_count + ?", delta)).
		Error
}

func (d *NoteDAO) Transaction(ctx context.Context, fn func(*gorm.DB) error) error {
	return d.Db.WithContext(ctx).Transaction(fn)
}

// GetNotesByTopicID - 根据话题ID获取相关笔记（按热度排序）
func (d *NoteDAO) GetNotesByTopicID(ctx context.Context, topicID uint64, limit, offset int) ([]*models.Note, error) {
	var notes []*models.Note
	err := d.Db.WithContext(ctx).
		Joins("INNER JOIN note_topics ON notes.id = note_topics.note_id").
		Where("note_topics.topic_id = ? AND notes.status = ?", topicID, 1).
		Order("notes.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notes).Error
	return notes, err
}

func (d *NoteDAO) ListNodeByUser(
	ctx context.Context,
	cursor int64,
	limit int,
	userId int,
) (notes []*models.Note, err error) {

	db := d.Db.WithContext(ctx).Model(&models.Note{})

	// 先限定用户
	db = db.Where("user_id = ?", userId)

	// 如果前端传了游标（大于0），则查询该时间点之前的数据
	if cursor > 0 {
		cursorTime := time.Unix(0, cursor)
		db = db.Where("created_at < ?", cursorTime)
	}

	err = db.
		Order("created_at DESC").
		Limit(limit).
		Find(&notes).Error

	return notes, err
}

func (d *NoteDAO) ListNodeByUserIDs(ctx context.Context, userIDs []int, cursor int64, limit int) ([]models.Note, error) {
	var nodes []models.Note
	query := d.Db.WithContext(ctx).Where("user_id IN ?", userIDs)

	if cursor > 0 {
		query = query.Where("id < ?", cursor)
	}
	err := query.Order("id DESC").Limit(limit).Find(&nodes).Error
	return nodes, err
}
