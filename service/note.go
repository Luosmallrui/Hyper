package service

import (
	"Hyper/dao"
	"Hyper/models"
	"Hyper/pkg/snowflake"
	"Hyper/types"
	"context"
	"encoding/json"
	"errors"
	"time"
)

var _ INoteService = (*NoteService)(nil)

type INoteService interface {
	CreateNote(ctx context.Context, userID uint64, req *types.CreateNoteRequest) (uint64, error)
	GetUserNotes(ctx context.Context, userID uint64, status int, limit, offset int) ([]*models.Note, error)
	UpdateNoteStatus(ctx context.Context, noteID uint64, status int) error
	ListNode(ctx context.Context, cursor int64, pageSize int) (types.ListNotesRep, error)
}
type NoteService struct {
	NoteDAO     *dao.NoteDAO
	UserService IUserService
}

func (s *NoteService) ListNode(ctx context.Context, cursor int64, pageSize int) (types.ListNotesRep, error) {
	limit := pageSize + 1
	nodes, err := s.NoteDAO.ListNode(ctx, cursor, limit)
	if err != nil {
		return types.ListNotesRep{}, err
	}

	rep := types.ListNotesRep{
		Notes:   make([]*types.Notes, 0),
		HasMore: false,
	}

	actualCount := len(nodes)
	if actualCount == 0 {
		return rep, nil
	}

	displayCount := actualCount
	if actualCount > pageSize {
		rep.HasMore = true
		displayCount = pageSize
	}

	// 只收集需要展示的数据的用户ID
	userIds := make([]uint64, 0, displayCount)
	for i := 0; i < displayCount; i++ {
		userIds = append(userIds, nodes[i].UserID)
	}
	userMap := s.UserService.BatchGetUserInfo(ctx, userIds)

	for i := 0; i < displayCount; i++ {
		note := nodes[i]
		dto := &types.Notes{
			ID:          int64(note.ID),
			UserID:      int64(note.UserID),
			Title:       note.Title,
			Content:     note.Content,
			Type:        note.Type,
			Status:      note.Status,
			VisibleConf: note.VisibleConf,
			CreatedAt:   note.CreatedAt,
			UpdatedAt:   note.UpdatedAt,
		}

		// 安全地获取用户信息
		if user, ok := userMap[note.UserID]; ok {
			dto.Avatar = user.Avatar
			dto.Nickname = user.Nickname
		}

		if err := json.Unmarshal([]byte(note.TopicIDs), &dto.TopicIDs); err != nil {
			dto.TopicIDs = make([]int64, 0)
		}
		if err := json.Unmarshal([]byte(note.Location), &dto.Location); err != nil {
			dto.Location = types.Location{}
		}

		var noteMedia []types.NoteMedia
		if err := json.Unmarshal([]byte(note.MediaData), &noteMedia); err == nil && len(noteMedia) > 0 {
			dto.MediaData = noteMedia[0]
		} else {
			dto.MediaData = types.NoteMedia{}
		}

		rep.Notes = append(rep.Notes, dto)
	}

	rep.NextCursor = nodes[displayCount-1].CreatedAt.UnixNano()

	return rep, nil
}

// CreateNote 创建笔记
func (s *NoteService) CreateNote(ctx context.Context, userID uint64, req *types.CreateNoteRequest) (uint64, error) {
	// 参数验证
	if req.Title == "" {
		return 0, errors.New("标题不能为空")
	}

	// 生成笔记ID
	noteID := uint64(snowflake.GenUserID())
	if len(req.TopicIDs) == 0 {
		req.TopicIDs = make([]int64, 0)
	}
	if len(req.MediaData) == 0 {
		req.MediaData = make([]types.NoteMedia, 0)
	}

	// 序列化 JSON 字段
	topicIDsJSON, err := json.Marshal(req.TopicIDs)
	if err != nil {
		return 0, err
	}

	// 修改这里: Location 为 nil 时使用 "{}" 或 "null"
	locationJSON := "{}" // 或者用 "null"
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
		Status:      0,
		VisibleConf: req.VisibleConf,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if note.VisibleConf == 0 {
		note.VisibleConf = types.VisibleConfPublic
	}

	// 保存到数据库
	if err := s.NoteDAO.Create(ctx, note); err != nil {
		return 0, err
	}

	return noteID, nil
}

// GetUserNotes 获取用户的笔记列表
func (s *NoteService) GetUserNotes(ctx context.Context, userID uint64, status int, limit, offset int) ([]*models.Note, error) {
	return s.NoteDAO.FindByUserID(ctx, userID, status, limit, offset)
}

// UpdateNoteStatus 更新笔记状态
func (s *NoteService) UpdateNoteStatus(ctx context.Context, noteID uint64, status int) error {
	return s.NoteDAO.UpdateStatus(ctx, noteID, status)
}
