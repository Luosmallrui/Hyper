package service

import (
	"Hyper/dao"
	"Hyper/models"
	"Hyper/pkg/snowflake"
	"Hyper/types"
	"context"
	"encoding/json"
	"context"
	"errors"
	"time"
)

var _ INoteService = (*NoteService)(nil)

type INoteService interface {
	CreateNote(ctx context.Context, userID uint64, req *types.CreateNoteRequest) (uint64, error)
	GetUserNotes(ctx context.Context, userID uint64, status int8, limit, offset int) ([]*models.Note, error)
	UpdateNoteStatus(ctx context.Context, noteID uint64, status int8) error
}

type NoteService struct {
	NoteDAO *dao.NoteDAO
}

// CreateNote 创建笔记
func (s *NoteService) CreateNote(ctx context.Context, userID uint64, req *types.CreateNoteRequest) (uint64, error) {
	// 参数验证
	if req.Title == "" {
		return 0, errors.New("标题不能为空")
	}

	// 生成笔记ID
	noteID := uint64(snowflake.GenUserID())

	// 序列化 JSON 字段
	topicIDsJSON, err := json.Marshal(req.TopicIDs)
	if err != nil {
		return 0, err
	}

	locationJSON := ""
	if req.Location != nil {
		locBytes, err := json.Marshal(req.Location)
		if err != nil {
			return 0, err
		}
		locationJSON = string(locBytes)
	}

	mediaDataJSON, err := json.Marshal(req.MediaData)
	if err != nil {
		return 0, err
	}

	// 构建笔记对象
	note := &models.Note{
		ID:          noteID,
		UserID:      userID,
		Title:       req.Title,
		Content:     req.Content,
		TopicIDs:    string(topicIDsJSON),
		Location:    locationJSON,
		MediaData:   string(mediaDataJSON),
		Type:        req.Type,
		Status:      0, // 默认审核中
		VisibleConf: req.VisibleConf,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 如果未指定可见性，默认为公开
	if note.VisibleConf == types.VisibleUnspecified {
		note.VisibleConf = types.VisiblePublic

	// 保存到数据库
	if err := s.NoteDAO.Create(ctx, note); err != nil {
		return 0, err
	}

	return noteID, nil
}

// GetUserNotes 获取用户的笔记列表
func (s *NoteService) GetUserNotes(ctx context.Context, userID uint64, status int8, limit, offset int) ([]*models.Note, error) {
	return nil
}

// UpdateNoteStatus 更新笔记状态
func (s *NoteService) UpdateNoteStatus(ctx context.Context, noteID uint64, status int8) error {
	return s.NoteDAO.UpdateStatus(ctx, noteID, status)
}
