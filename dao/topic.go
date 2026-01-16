package dao

import (
	"Hyper/models"
	"context"

	"gorm.io/gorm"
)

type Topic struct {
	Repo[models.Topic]
}

type NoteTopic struct {
	Repo[models.NoteTopic]
}

func NewNoteTopic(db *gorm.DB) *NoteTopic {
	return &NoteTopic{
		Repo: NewRepo[models.NoteTopic](db),
	}
}

func NewTopic(db *gorm.DB) *Topic {
	return &Topic{
		Repo: NewRepo[models.Topic](db),
	}
}

// 创建话题
func (d *Topic) CreateTopic(ctx context.Context, topic *models.Topic) error {
	return d.Db.WithContext(ctx).Create(topic).Error
}

// GetHotTopics - 获取热门话题（按热度和最后发帖时间排序）
func (d *Topic) GetHotTopics(ctx context.Context, limit int) ([]*models.Topic, error) {
	var topics []*models.Topic
	err := d.Db.WithContext(ctx).
		Where("status = ?", 1). // status=1 表示正常
		Order("is_hot DESC, sort_weight DESC, last_post_at DESC").
		Limit(limit).
		Find(&topics).Error
	return topics, err
}

// FindTopicsByName - 按名称模糊搜索话题（按热度排序）
func (d *Topic) FindTopicsByName(ctx context.Context, keyword string, limit int) ([]*models.Topic, error) {
	var topics []*models.Topic
	err := d.Db.WithContext(ctx).
		Where("status = ? AND name LIKE ?", 1, "%"+keyword+"%").
		Order("view_count DESC, sort_weight DESC, last_post_at DESC").
		Limit(limit).
		Find(&topics).Error
	return topics, err
}

// FindTopicByName - 根据名称精确查询话题
func (d *Topic) FindTopicByName(ctx context.Context, name string) (*models.Topic, error) {
	var topic *models.Topic
	err := d.Db.WithContext(ctx).
		Where("status = ? AND name = ?", 1, name).
		First(&topic).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil // 不存在返回 nil
	}
	return topic, err
}
func (d *Topic) Transaction(ctx context.Context, f func(tx *gorm.DB) error) interface{} {
	return d.Db.WithContext(ctx).Transaction(f)
}

// CreateNoteTopic - 创建笔记与话题的关联
func (d *NoteTopic) CreateNoteTopic(ctx context.Context, noteTopic *models.NoteTopic) error {
	return d.Db.WithContext(ctx).Create(noteTopic).Error
}

// 批量创建笔记与话题的关联
func (d *NoteTopic) BatchCreateNoteTopic(ctx context.Context, noteTopics []*models.NoteTopic) error {
	if len(noteTopics) == 0 {
		return nil
	}
	return d.Db.WithContext(ctx).Create(&noteTopics).Error
}

// 根据笔记ID删除关联的话题
func (d *NoteTopic) DeleteTopicByNoteID(ctx context.Context, noteID uint64) error {
	return d.Db.WithContext(ctx).Where("note_id = ?", noteID).Delete(&models.NoteTopic{}).Error
}

// 获取笔记关联的所有话题
func (d *NoteTopic) GetTopicsByNoteID(ctx context.Context, noteID uint64) ([]uint64, error) {
	var topicIDs []uint64
	err := d.Db.WithContext(ctx).
		Where("note_id = ?", noteID).
		Pluck("topic_id", &topicIDs).Error
	return topicIDs, err
}

// GetNotesByTopic - 根据话题ID获取相关笔记（支持分页）
func (d *Topic) GetNotesByTopic(ctx context.Context, topicID uint64, limit, offset int) ([]*models.Note, error) {
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

// GetTopicByID - 根据ID获取话题
func (d *Topic) GetTopicByID(ctx context.Context, topicID uint64) (*models.Topic, error) {
	var topic *models.Topic
	err := d.Db.WithContext(ctx).
		Where("id = ? AND status = 1", topicID).
		First(&topic).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return topic, err
}
