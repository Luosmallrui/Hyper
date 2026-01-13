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
