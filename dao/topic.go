package dao

import (
	"Hyper/models"
	"context"
	"gorm.io/gorm"
	"time"
)

type Topic struct {
	Repo[models.Topic]
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

func (d *Topic) Transaction(ctx context.Context, f func(tx *gorm.DB) error) interface{} {
	return d.Db.WithContext(ctx).Transaction(f)
}

func (d *Topic) FindTopicByName(ctx context.Context, name string) (*models.Topic, error) {
	var topic models.Topic

	err := d.Db.WithContext(ctx).
		Model(&models.Topic{}).
		Where("name = ? AND status = 1", name).
		First(&topic).Error

	if err != nil {
		return nil, err
	}
	return &topic, nil
}

// 批量根据话题名称获取话题
func (d *Topic) BatchTopicFindByNames(ctx context.Context, names []string) (map[string]*models.Topic, error) {
	var topics []*models.Topic

	err := d.Db.WithContext(ctx).
		Where("name IN ? AND status = 1", names).
		Find(&topics).Error

	if err != nil {
		return nil, err
	}

	// 转换为 map，方便查找
	topicMap := make(map[string]*models.Topic)
	for _, topic := range topics {
		topicMap[topic.Name] = topic
	}

	return topicMap, nil
}

func (d *Topic) BatchTopicFindByIDs(ctx context.Context, topicIDs []uint64) (map[uint64]*models.Topic, error) {
	var topics []*models.Topic

	err := d.Db.WithContext(ctx).
		Where("id IN ? AND status = 1", topicIDs).
		Find(&topics).Error

	if err != nil {
		return nil, err
	}

	topicMap := make(map[uint64]*models.Topic)
	for _, topic := range topics {
		topicMap[topic.ID] = topic
	}

	return topicMap, nil
}

func (d *Topic) FindTopicByID(ctx context.Context, topicID uint64) (*models.Topic, error) {
	var topic models.Topic
	err := d.Db.WithContext(ctx).
		Where("id = ? AND status = 1", topicID).
		First(&topic).Error
	if err != nil {
		return nil, err
	}
	return &topic, nil
}

// 批量创建笔记与话题的关联
func (d *Topic) BatchCreateNoteTopicAssociations(ctx context.Context, noteID uint64, topicIDs []uint64) error {
	if len(topicIDs) == 0 {
		return nil
	}

	associations := make([]*models.NoteTopic, 0, len(topicIDs))
	now := time.Now()

	for _, topicID := range topicIDs {
		associations = append(associations, &models.NoteTopic{
			NoteID:    noteID,
			TopicID:   topicID,
			CreatedAt: now,
		})
	}

	return d.Db.WithContext(ctx).CreateInBatches(associations, 100).Error
}

// 批量增加话题的发布数
func (d *Topic) BatchUpdateTopicsUpdatedAt(ctx context.Context, topicIDs []uint64, delta int32) error {
	if len(topicIDs) == 0 {
		return nil
	}

	return d.Db.WithContext(ctx).
		Model(&models.Topic{}).
		Where("id IN ?", topicIDs).
		Updates(map[string]interface{}{
			// 使用 GREATEST 函数确保计数值不会低于 0
			"post_count":   gorm.Expr("GREATEST(post_count + ?, 0)", delta),
			"last_post_at": time.Now(),
		}).Error
}

func (d *Topic) GetNoteIDsByTopicWithCursor(ctx context.Context, topicID uint64, cursor int64, pageSize int) ([]uint64, int64, bool, error) {
	var noteTopics []*models.NoteTopic

	limit := pageSize + 1 // 多查一条判断是否有更多

	query := d.Db.WithContext(ctx).
		Table("note_topics nt").
		Select("nt.note_id, n.created_at").
		Joins("JOIN notes n ON nt.note_id = n.id").
		Joins("JOIN note_stats ns ON nt.note_id = ns.note_id").
		Where("nt.topic_id = ? AND n.status = 1", topicID)

	// Cursor 逻辑：根据 (创建时间, noteID) 排序，支持游标
	if cursor > 0 {
		query = query.Where("n.created_at < ?", time.UnixMicro(cursor/1000))
		// 或者：
		query = query.Where("n.created_at < ?", time.Unix(0, cursor))
	}

	err := query.
		Order("n.created_at DESC, nt.note_id DESC").
		Limit(limit).
		Scan(&noteTopics).Error

	if err != nil {
		return nil, 0, false, err
	}

	// 提取笔记ID
	noteIDs := make([]uint64, 0)
	hasMore := false
	var nextCursor int64

	displayCount := len(noteTopics)
	if displayCount > pageSize {
		hasMore = true
		displayCount = pageSize
	}

	for i := 0; i < displayCount; i++ {
		noteIDs = append(noteIDs, noteTopics[i].NoteID)
	}

	// 计算下一个 Cursor（最后一条记录的创建时间）
	if displayCount > 0 {
		lastTopic := noteTopics[displayCount-1]
		nextCursor = lastTopic.CreatedAt.UnixNano()
	}

	return noteIDs, nextCursor, hasMore, nil
}

func (d *Topic) BatchCreateTopics(ctx context.Context, topics []*models.Topic) error {
	if len(topics) == 0 {
		return nil
	}
	return d.Db.WithContext(ctx).CreateInBatches(topics, 100).Error
}
