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

// NoteClassifySubscribe 消费笔记分类消息：调用大模型给未分类笔记打频道标签。
type NoteClassifySubscribe struct {
	DB             *gorm.DB
	ChannelService service.IChannelService
}

func (m *NoteClassifySubscribe) Init() error { return nil }

func (m *NoteClassifySubscribe) Setup(ctx context.Context) error { return nil }

func (m *NoteClassifySubscribe) handleNoteClassify(ctx context.Context, mv *rmq_client.MessageView) error {
	var msg types.NoteClassifyMessage
	if err := json.Unmarshal(mv.GetBody(), &msg); err != nil || msg.NoteID == 0 {
		if err != nil {
			log.L.Error("unmarshal note classify msg error", zap.Error(err))
		}
		return nil // 消息格式错误，重试无意义，直接 ack 丢弃
	}

	// 幂等：只处理 channel_id = 0 的笔记，已分类/不存在则跳过
	var note models.Note
	err := m.DB.WithContext(ctx).Where("id = ? AND channel_id = ?", msg.NoteID, 0).First(&note).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // 已分类或不存在，跳过
		}
		return err // DB 错误，重试
	}

	channels, err := m.ChannelService.ListChannels(ctx, &types.ListChannelsReq{})
	if err != nil || channels == nil || len(channels.Channels) == 0 {
		return err // 拉不到频道列表，重试
	}
	tagsSlice, tagsMap := buildChannelMap(channels.Channels)

	// LLM 调用加超时，避免模型侧长时间无响应占住消费协程（与 note_image_tag.go 对齐）
	llmCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	label := llm.ClassifyMultiImageNote(llmCtx, note.Title, note.Content, noteImageURLs(note.MediaData), tagsSlice)
	cancel()
	labelID, ok := tagsMap[label]
	if !ok || labelID == 0 {
		// 分类失败/不匹配：不重试，避免无意义计费
		log.L.Warn("note classify no match", zap.Uint64("note_id", note.ID), zap.String("label", label))
		return nil
	}

	if err := m.DB.Model(&models.Note{}).Where("id = ?", note.ID).Update("channel_id", labelID).Error; err != nil {
		return err // 更新失败，重试
	}
	return nil
}

// buildChannelMap 由频道列表构建名称切片与名称→ID 映射。
func buildChannelMap(channels []*types.ChannelInfo) ([]string, map[string]int) {
	tagsSlice := make([]string, 0, len(channels))
	tagsMap := make(map[string]int, len(channels))
	for _, v := range channels {
		tagsSlice = append(tagsSlice, v.Name)
		tagsMap[v.Name] = v.Id
	}
	return tagsSlice, tagsMap
}

// noteImageURLs 解析笔记媒体数据，返回图片 URL 列表（分类时内部只取第一张）。
func noteImageURLs(mediaData string) []string {
	var noteMedia []types.NoteMedia
	if err := json.Unmarshal([]byte(mediaData), &noteMedia); err != nil {
		return nil
	}
	urls := make([]string, 0, len(noteMedia))
	for _, m := range noteMedia {
		if m.URL != "" {
			urls = append(urls, m.URL)
		}
	}
	return urls
}
