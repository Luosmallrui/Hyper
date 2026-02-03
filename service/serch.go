package service

import (
	"Hyper/config"
	"Hyper/models"
	"Hyper/types"
	"context"
	"errors"
	"sync"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

type SearchService struct {
	Config *config.Config
	DB     *gorm.DB
	Redis  *redis.Client
}

var _ ISearchService = (*SearchService)(nil)

type ISearchService interface {
	GlobalSerch(ctx context.Context, req types.GlobalSearchReq) (*types.GlobalSearchResp, error)
}

func (s *SearchService) GlobalSerch(ctx context.Context, req types.GlobalSearchReq) (*types.GlobalSearchResp, error) {
	// 防御性检查
	if s.DB == nil {
		return nil, errors.New("database connection is nil")
	}

	var (
		g    errgroup.Group
		resp = &types.GlobalSearchResp{}

		mu sync.Mutex

		dbUsers   []models.Users
		dbNotes   []models.Note
		dbParties []models.Merchant
	)

	keyword := "%" + req.Keyword + "%"

	if (req.Type == 0 && req.NoteCursor == 0) || req.Type == 1 {
		g.Go(func() error {
			db := s.DB.WithContext(ctx).Model(&models.Users{}).Where("nickname LIKE ? OR id LIKE ?", keyword, keyword)
			if req.UserCursor > 0 {
				db = db.Where("id < ?", req.UserCursor)
			}
			limit := 3
			if req.Type == 1 {
				limit = req.Limit
			}

			err := db.Order("id DESC").Limit(limit).Find(&dbUsers).Error
			if err != nil {
				return err
			}

			if len(dbUsers) > 0 {
				mu.Lock()
				resp.NextUserCursor = uint64(dbUsers[len(dbUsers)-1].Id)
				mu.Unlock()
			}
			return nil
		})
	}

	if req.Type == 0 || req.Type == 2 {
		g.Go(func() error {
			db := s.DB.WithContext(ctx).Model(&models.Note{}).
				Where("(title LIKE ? OR content LIKE ?) AND status = 0 ", keyword, keyword)
			if req.NoteCursor > 0 {
				db = db.Where("id < ?", req.NoteCursor)
			}
			limit := req.Limit
			if req.Type == 0 {
				limit = 5
			}

			err := db.Order("id DESC").Limit(limit).Find(&dbNotes).Error
			if err != nil {
				return err
			}

			if len(dbNotes) > 0 {
				mu.Lock()
				resp.NextNoteCursor = dbNotes[len(dbNotes)-1].ID
				mu.Unlock()
			}
			return nil
		})
	}

	if req.Type == 0 || req.Type == 3 {
		g.Go(func() error {
			db := s.DB.WithContext(ctx).Model(&models.Merchant{}).
				Where("(title LIKE ? OR location_name LIKE ?) AND status IN (0,1)", keyword, keyword)
			if req.PartyCursor > 0 {
				db = db.Where("id < ?", req.PartyCursor)
			}

			err := db.Order("id DESC").Limit(req.Limit).Find(&dbParties).Error
			if err != nil {
				return err
			}

			if len(dbParties) > 0 {
				mu.Lock()
				resp.NextPartyCursor = dbParties[len(dbParties)-1].ID
				mu.Unlock()
			}
			return nil
		})
	}

	// 等待所有查询
	if err := g.Wait(); err != nil {
		return nil, err
	}

	if len(dbUsers) > 0 {
		resp.Users = make([]types.SearchUserItem, 0, len(dbUsers))
		for _, u := range dbUsers {
			resp.Users = append(resp.Users, types.SearchUserItem{
				ID:       int(u.Id),
				Nickname: u.Nickname,
				Avatar:   u.Avatar,
				Motto:    u.Motto,
			})
		}
	}

	if len(dbNotes) > 0 {
		userIds := make([]int, 0)
		for _, n := range dbNotes {
			userIds = append(userIds, int(n.UserID))
		}
		userMap := s.getUsersMap(ctx, userIds) // 辅助方法

		resp.Notes = make([]types.SearchNoteItem, 0, len(dbNotes))
		for _, n := range dbNotes {
			item := types.SearchNoteItem{
				ID:        n.ID,
				Title:     n.Title,
				Content:   n.Content,
				Type:      n.Type,
				CreatedAt: n.CreatedAt,
				Cover:     n.MediaData,
			}
			if u, ok := userMap[int(n.UserID)]; ok {
				item.User = u
			}
			resp.Notes = append(resp.Notes, item)
		}
	}
	// 3. 组装活动
	if len(dbParties) > 0 {
		resp.Parties = make([]types.SearchPartyItem, 0, len(dbParties))
		for _, p := range dbParties {
			resp.Parties = append(resp.Parties, types.SearchPartyItem{
				ID:           p.ID,
				Title:        p.Title,
				Type:         p.Type,
				LocationName: p.LocationName,
			})
		}
	}
	return resp, nil
}

func (s *SearchService) getUsersMap(ctx context.Context, userIds []int) map[int]types.SearchUserItem {
	var users []models.Users
	s.DB.WithContext(ctx).Where("id IN ?", userIds).Find(&users)

	res := make(map[int]types.SearchUserItem)
	for _, u := range users {
		res[u.Id] = types.SearchUserItem{
			ID:       u.Id,
			Nickname: u.Nickname,
			Avatar:   u.Avatar,
		}
	}
	return res
}
