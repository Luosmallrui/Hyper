package service

import (
	"Hyper/config"
	"Hyper/models"
	"Hyper/types"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

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
	SaveSearchHistory(ctx context.Context, userId int, keyword string) error
	GetSearchHistory(ctx context.Context, userId int) ([]string, error)
	DeleteSearchKeyword(ctx context.Context, userId int, keyword string) error
}

func (s *SearchService) DeleteSearchKeyword(ctx context.Context, userId int, keyword string) error {
	key := fmt.Sprintf("search:history:%d", userId)
	return s.Redis.ZRem(ctx, key, keyword).Err()
}

func (s *SearchService) GetSearchHistory(ctx context.Context, userId int) ([]string, error) {
	key := fmt.Sprintf("search:history:%d", userId)

	result, err := s.Redis.ZRevRange(ctx, key, 0, 9).Result()
	if err != nil {
		return nil, err
	}

	return result, nil
}
func (s *SearchService) SaveSearchHistory(ctx context.Context, userId int, keyword string) error {
	key := fmt.Sprintf("search:history:%d", userId)
	now := float64(time.Now().Unix())

	// 先删除旧的（去重）
	s.Redis.ZRem(ctx, key, keyword)

	// 添加新的
	err := s.Redis.ZAdd(ctx, key, redis.Z{
		Score:  now,
		Member: keyword,
	}).Err()
	if err != nil {
		return err
	}

	// 只保留最新 10 条
	s.Redis.ZRemRangeByRank(ctx, key, 0, -11)

	// 设置过期时间（可选）
	s.Redis.Expire(ctx, key, 30*24*time.Hour)

	return nil
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

		dbUsers      []models.Users
		dbNotes      []models.Note
		dbParties    []models.Merchant
		dbActivities []models.Activity
	)

	keyword := "%" + strings.TrimSpace(req.Keyword) + "%"

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
			db := s.DB.WithContext(ctx).Model(&models.Merchant{}).Where("status = ?", "active")
			if strings.TrimSpace(req.Keyword) != "" {
				db = db.Where("(title LIKE ? OR location_name LIKE ? OR address LIKE ?)", keyword, keyword, keyword)
			}
			if req.District != "" {
				db = db.Where("district_id = ?", req.District)
			}
			if req.Area != "" {
				db = db.Where("area_id = ?", req.Area)
			}
			if req.BusinessArea != "" {
				db = db.Where("location_name LIKE ? OR address LIKE ?", "%"+req.BusinessArea+"%", "%"+req.BusinessArea+"%")
			}
			if tagBits := parseSearchTagBits(req.Tags); tagBits > 0 {
				db = db.Where("tags & ? = ?", tagBits, tagBits)
			}
			if req.Distance > 0 && req.Lat != 0 && req.Lng != 0 {
				db = db.Where(searchDistanceSQL("latitude", "longitude")+" <= ?", req.Lat, req.Lng, req.Lat, req.Distance)
			}
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

	if req.Type == 0 || req.Type == 4 {
		g.Go(func() error {
			db := s.DB.WithContext(ctx).Model(&models.Activity{}).Where("status = ?", models.ActivityStatusOnline)
			if strings.TrimSpace(req.Keyword) != "" {
				db = db.Where("(name LIKE ? OR address LIKE ? OR province LIKE ? OR city LIKE ? OR district LIKE ?)", keyword, keyword, keyword, keyword, keyword)
			}
			if req.District != "" {
				db = db.Where("district = ?", req.District)
			}
			if req.Area != "" {
				db = db.Where("address LIKE ?", "%"+req.Area+"%")
			}
			if req.BusinessArea != "" {
				db = db.Where("address LIKE ?", "%"+req.BusinessArea+"%")
			}
			if req.Distance > 0 && req.Lat != 0 && req.Lng != 0 {
				db = db.Where(searchDistanceSQL("latitude", "longitude")+" <= ?", req.Lat, req.Lng, req.Lat, req.Distance)
			}
			if req.PartyCursor > 0 {
				db = db.Where("id < ?", req.PartyCursor)
			}
			if err := db.Order("id DESC").Limit(req.Limit).Find(&dbActivities).Error; err != nil {
				return err
			}
			if len(dbActivities) > 0 {
				mu.Lock()
				resp.NextPartyCursor = dbActivities[len(dbActivities)-1].ID
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
				CoverImage:   p.CoverImage,
				Price:        76,
				StartTime:    p.CreatedAt,
			})
		}
	}
	if len(dbActivities) > 0 {
		activityIDs := make([]int64, 0, len(dbActivities))
		for _, a := range dbActivities {
			activityIDs = append(activityIDs, a.ID)
		}
		priceMap := s.activityMinPriceMap(ctx, activityIDs)
		resp.Activities = make([]types.SearchActivityItem, 0, len(dbActivities))
		for _, a := range dbActivities {
			resp.Activities = append(resp.Activities, types.SearchActivityItem{
				ID:         a.ID,
				Name:       a.Name,
				PosterList: a.PosterList,
				StartTime:  a.StartTime,
				EndTime:    a.EndTime,
				Province:   a.Province,
				City:       a.City,
				District:   a.District,
				Address:    a.Address,
				Lat:        a.Latitude,
				Lng:        a.Longitude,
				AvgPrice:   priceMap[a.ID],
				Status:     a.Status,
			})
		}
	}
	return resp, nil
}

func parseSearchTagBits(raw string) int {
	var bits int
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		var id int
		_, _ = fmt.Sscanf(item, "%d", &id)
		if id > 0 {
			bits |= id
		}
	}
	return bits
}

func searchDistanceSQL(latColumn, lngColumn string) string {
	return fmt.Sprintf(`(6371 * acos(
		cos(radians(?)) *
		cos(radians(%s)) *
		cos(radians(%s) - radians(?)) +
		sin(radians(?)) *
		sin(radians(%s))
	))`, latColumn, lngColumn, latColumn)
}

func (s *SearchService) activityMinPriceMap(ctx context.Context, activityIDs []int64) map[int64]int64 {
	result := make(map[int64]int64)
	if len(activityIDs) == 0 {
		return result
	}
	var rows []struct {
		ActivityID int64
		MinPrice   int64
	}
	if err := s.DB.WithContext(ctx).Model(&models.TicketSpec{}).
		Select("activity_id, MIN(price) AS min_price").
		Where("activity_id IN ?", activityIDs).
		Group("activity_id").
		Scan(&rows).Error; err != nil {
		return result
	}
	for _, row := range rows {
		result[row.ActivityID] = row.MinPrice
	}
	return result
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
