package process

import (
	"Hyper/models"
	"Hyper/pkg/llm"
	"Hyper/pkg/log"
	"Hyper/service"
	"Hyper/types"
	"context"
	"encoding/json"
	"errors"
	"time"

	rmq_client "github.com/apache/rocketmq-clients/golang/v5"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// NoteImageTagSubscribe 消费图片标签任务。上传接口只入队，这里再调用视觉 LLM。
type NoteImageTagSubscribe struct {
	DB               *gorm.DB
	TopicService     service.ITopicService
	MessageSubscribe *MessageSubscribe
}

func NewNoteImageTagSubscribe(db *gorm.DB, messageSubscribe *MessageSubscribe) *NoteImageTagSubscribe {
	return &NoteImageTagSubscribe{
		DB:               db,
		TopicService:     &service.TopicService{DB: db},
		MessageSubscribe: messageSubscribe,
	}
}

func (m *NoteImageTagSubscribe) Init() error { return nil }

func (m *NoteImageTagSubscribe) Setup(ctx context.Context) error { return nil }

func (m *NoteImageTagSubscribe) handleNoteImageTag(ctx context.Context, mv *rmq_client.MessageView) error {
	var msg types.NoteImageTagMessage
	if err := json.Unmarshal(mv.GetBody(), &msg); err != nil || msg.ImageID <= 0 || msg.UserID == 0 || msg.URL == "" {
		if err != nil {
			log.L.Error("unmarshal note image tag msg error", zap.Error(err))
		}
		return nil // 非法任务重试无意义，直接 Ack
	}
	log.L.Info("consume note image tag task", zap.Int64("image_id", msg.ImageID), zap.Uint64("user_id", msg.UserID))

	var image models.Image
	if err := m.DB.WithContext(ctx).Where("id = ? AND user_id = ?", msg.ImageID, msg.UserID).First(&image).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	if image.TagStatus == types.ImageTagStatusCompleted {
		return m.pushStoredResult(ctx, msg.UserID, image, msg.URL)
	}

	tagCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	tags := llm.GenNoteTag(tagCtx, msg.URL)
	cancel()
	if len(tags) == 0 {
		log.L.Warn("note image tag generation returned no tags", zap.Int64("image_id", msg.ImageID))
		return m.finishFailed(ctx, msg, "标签生成失败")
	}

	topicTags, err := m.TopicService.GetOrCreateTopicIDs(ctx, tags, msg.UserID)
	if err != nil {
		return err // 话题写入失败由 RocketMQ 重试
	}
	encodedTags, err := json.Marshal(topicTags)
	if err != nil {
		return err
	}
	if err := m.DB.WithContext(ctx).Model(&models.Image{}).Where("id = ?", image.ID).
		Updates(map[string]interface{}{
			"tags":       string(encodedTags),
			"tag_status": types.ImageTagStatusCompleted,
			"tag_error":  "",
		}).Error; err != nil {
		return err
	}

	m.pushResult(ctx, msg.UserID, types.NoteImageTagResult{
		ImageID: image.ID,
		URL:     msg.URL,
		Status:  "completed",
		Tags:    topicTags,
	})
	log.L.Info("note image tag task completed", zap.Int64("image_id", image.ID), zap.Int("tag_count", len(topicTags)))
	return nil
}

func (m *NoteImageTagSubscribe) finishFailed(ctx context.Context, msg types.NoteImageTagMessage, reason string) error {
	if err := m.DB.WithContext(ctx).Model(&models.Image{}).Where("id = ? AND user_id = ?", msg.ImageID, msg.UserID).
		Updates(map[string]interface{}{
			"tag_status": types.ImageTagStatusFailed,
			"tag_error":  reason,
		}).Error; err != nil {
		return err
	}
	m.pushResult(ctx, msg.UserID, types.NoteImageTagResult{
		ImageID: msg.ImageID,
		URL:     msg.URL,
		Status:  "failed",
		Tags:    make([]types.CreateOrGetTopicResponse, 0),
		Error:   "标签生成失败",
	})
	log.L.Warn("note image tag task failed", zap.Int64("image_id", msg.ImageID), zap.String("reason", reason))
	return nil
}

func (m *NoteImageTagSubscribe) pushStoredResult(ctx context.Context, userID uint64, image models.Image, url string) error {
	tags := make([]types.CreateOrGetTopicResponse, 0)
	if image.Tags != "" {
		if err := json.Unmarshal([]byte(image.Tags), &tags); err != nil {
			return err
		}
	}
	m.pushResult(ctx, userID, types.NoteImageTagResult{
		ImageID: image.ID,
		URL:     url,
		Status:  "completed",
		Tags:    tags,
	})
	return nil
}

func (m *NoteImageTagSubscribe) pushResult(ctx context.Context, userID uint64, result types.NoteImageTagResult) {
	if m.MessageSubscribe == nil {
		return
	}
	pushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	m.MessageSubscribe.PushEventToUser(pushCtx, int(userID), "note.image_tags", result)
}
