package handler

import (
	"Hyper/models"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Map struct {
	MapService service.IMapService
	OssService service.IOssService
	Redis      *redis.Client
	DB         *gorm.DB
}

func (m *Map) RegisterRouter(r gin.IRouter) {
	mapGroup := r.Group("/v1/districts")
	mapGroup.GET("/map", context.Wrap(m.GetMap))
	mapGroup.GET("/test", context.Wrap(m.Test))
	mapGroup.GET("/tree", context.Wrap(m.GetDistrictTree))

	markerGroup := r.Group("/v1/map")
	markerGroup.GET("/markers", context.Wrap(m.GetMarkers))
}

func (m *Map) Test(c *gin.Context) error {
	mapData, err := m.OssService.ListBuckets(c.Request.Context())
	if err != nil {
		return err
	}
	fmt.Println(m.Redis, 55)
	response.Success(c, mapData)
	return nil
}

func (m *Map) GetMap(c *gin.Context) error {
	mapData, err := m.MapService.GetMapData()
	if err != nil {
		response.Fail(c, 500, "获取地图数据失败")
		return err
	}
	response.Success(c, mapData)
	return nil
}

func (m *Map) GetMarkers(c *gin.Context) error {
	source := c.DefaultQuery("source", "all")
	if source != "all" && source != "party" && source != "activity" && source != "venue" && source != "merchant" {
		return errors.New("source 仅支持 all、party、activity、venue、merchant")
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "200"))
	if limit <= 0 {
		limit = 200
	}
	if limit > 500 {
		limit = 500
	}

	markers := make([]types.MapMarker, 0)
	if source == "all" || source == "party" || source == "venue" || source == "merchant" {
		parties, err := m.getPartyMarkers(c, limit)
		if err != nil {
			return err
		}
		markers = append(markers, parties...)
	}
	if source == "all" || source == "activity" {
		activities, err := m.getActivityMarkers(c, limit)
		if err != nil {
			return err
		}
		markers = append(markers, activities...)
	}

	response.Success(c, types.MapMarkerResponse{
		List:  markers,
		Total: len(markers),
	})
	return nil
}

func (m *Map) getPartyMarkers(c *gin.Context, limit int) ([]types.MapMarker, error) {
	var parties []models.Merchant
	query := m.DB.WithContext(c.Request.Context()).
		Where("status = ?", "active").
		Where("latitude <> 0 AND longitude <> 0")
	if categoryID := queryInt(c, "category_id"); categoryID > 0 {
		query = query.Where("category = ?", categoryID)
	}
	if districtID := queryInt(c, "district_id"); districtID > 0 {
		query = query.Where("district_id = ?", districtID)
	}
	if areaID := queryInt(c, "area_id"); areaID > 0 {
		query = query.Where("area_id = ?", areaID)
	}
	if tagBits := parseTagBits(c.Query("tag_ids")); tagBits > 0 {
		query = query.Where("tags & ? = ?", tagBits, tagBits)
	}
	if err := query.Order("created_at DESC").Limit(limit).Find(&parties).Error; err != nil {
		return nil, err
	}

	userIDs := make([]int, 0, len(parties))
	for _, party := range parties {
		userIDs = append(userIDs, party.UserID)
	}
	userMap := make(map[int]models.Users)
	if len(userIDs) > 0 {
		var users []models.Users
		if err := m.DB.WithContext(c.Request.Context()).Where("id IN ?", userIDs).Find(&users).Error; err != nil {
			return nil, err
		}
		for _, user := range users {
			userMap[user.Id] = user
		}
	}

	markers := make([]types.MapMarker, 0, len(parties))
	requestedSource := c.DefaultQuery("source", "all")
	for _, party := range parties {
		icon := "https://cdn.hypercn.cn/icon/party.png"
		source := "merchant"
		if party.Type == "场地" {
			icon = "https://cdn.hypercn.cn/icon/jiuba.png"
			source = "venue"
		}
		if requestedSource == "venue" && source != "venue" {
			continue
		}
		if requestedSource == "merchant" && source != "merchant" {
			continue
		}
		marker := types.MapMarker{
			ID:           fmt.Sprintf("party-%d", party.ID),
			Source:       source,
			SourceID:     party.ID,
			UserID:       int64(party.UserID),
			Title:        party.Title,
			Type:         party.Type,
			Location:     party.LocationName,
			Address:      party.Address,
			Lat:          party.Latitude,
			Lng:          party.Longitude,
			CoverImage:   party.CoverImage,
			CreatedAt:    formatMarkerTime(party.CreatedAt),
			AvgPrice:     7600,
			CurrentCount: 9932,
			PostCount:    372,
			Icon:         icon,
			Status:       party.Status,
			CategoryID:   party.Category,
			DistrictID:   party.DistrictID,
			AreaID:       party.AreaID,
			TagIDs:       tagBitsToIDs(party.Tags),
		}
		if user, ok := userMap[party.UserID]; ok {
			marker.User = user.Nickname
			marker.UserName = user.Nickname
			marker.UserAvatar = user.Avatar
			marker.UserAvatarCamel = user.Avatar
		}
		markers = append(markers, marker)
	}
	return markers, nil
}

func (m *Map) getActivityMarkers(c *gin.Context, limit int) ([]types.MapMarker, error) {
	var activities []models.Activity
	query := m.DB.WithContext(c.Request.Context()).
		Where("status = ?", models.ActivityStatusOnline).
		Where("latitude <> 0 AND longitude <> 0")
	if err := query.Order("created_at DESC").Limit(limit).Find(&activities).Error; err != nil {
		return nil, err
	}

	organizerIDs := make([]int64, 0, len(activities))
	activityIDs := make([]int64, 0, len(activities))
	for _, activity := range activities {
		organizerIDs = append(organizerIDs, activity.OrganizerID)
		activityIDs = append(activityIDs, activity.ID)
	}

	organizerMap := make(map[int64]models.Organizer)
	userIDs := make([]int, 0, len(organizerIDs))
	if len(organizerIDs) > 0 {
		var organizers []models.Organizer
		if err := m.DB.WithContext(c.Request.Context()).Where("id IN ?", organizerIDs).Find(&organizers).Error; err != nil {
			return nil, err
		}
		for _, organizer := range organizers {
			organizerMap[organizer.ID] = organizer
			userIDs = append(userIDs, int(organizer.UserID))
		}
	}

	userMap := make(map[int]models.Users)
	if len(userIDs) > 0 {
		var users []models.Users
		if err := m.DB.WithContext(c.Request.Context()).Where("id IN ?", userIDs).Find(&users).Error; err != nil {
			return nil, err
		}
		for _, user := range users {
			userMap[user.Id] = user
		}
	}

	priceMap := make(map[int64]int64)
	if len(activityIDs) > 0 {
		var rows []struct {
			ActivityID int64
			MinPrice   int64
		}
		if err := m.DB.WithContext(c.Request.Context()).Model(&models.TicketSpec{}).
			Select("activity_id, MIN(price) AS min_price").
			Where("activity_id IN ?", activityIDs).
			Group("activity_id").
			Scan(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			priceMap[row.ActivityID] = row.MinPrice
		}
	}

	markers := make([]types.MapMarker, 0, len(activities))
	for _, activity := range activities {
		marker := types.MapMarker{
			ID:           fmt.Sprintf("activity-%d", activity.ID),
			Source:       "activity",
			SourceID:     activity.ID,
			Title:        activity.Name,
			Type:         "activity",
			Location:     activity.Address,
			Address:      activity.Address,
			Lat:          activity.Latitude,
			Lng:          activity.Longitude,
			CoverImage:   activity.PosterList,
			CreatedAt:    formatMarkerTime(activity.CreatedAt),
			AvgPrice:     priceMap[activity.ID],
			CurrentCount: 0,
			PostCount:    0,
			Icon:         "https://cdn.hypercn.cn/icon/party.png",
			StartTime:    formatMarkerTime(activity.StartTime),
			EndTime:      formatMarkerTime(activity.EndTime),
			Status:       activity.Status,
		}
		if organizer, ok := organizerMap[activity.OrganizerID]; ok {
			marker.UserID = organizer.UserID
			marker.User = organizer.Name
			marker.UserName = organizer.Name
			marker.UserAvatar = organizer.Logo
			marker.UserAvatarCamel = organizer.Logo
			if user, ok := userMap[int(organizer.UserID)]; ok {
				if user.Nickname != "" {
					marker.User = user.Nickname
					marker.UserName = user.Nickname
				}
				if user.Avatar != "" {
					marker.UserAvatar = user.Avatar
					marker.UserAvatarCamel = user.Avatar
				}
			}
		}
		markers = append(markers, marker)
	}
	return markers, nil
}

func (m *Map) GetDistrictTree(c *gin.Context) error {
	var districts []models.District
	//city_id 暂时写死为1，后续可以增加城市选择功能
	if err := m.DB.WithContext(c.Request.Context()).
		Where("city_id = ?", 1).
		Order("sort_order asc").
		Find(&districts).Error; err != nil {
		return errors.New("获取行政区列表失败: " + err.Error())
	}

	var dbareas []models.Areas
	if err := m.DB.WithContext(c.Request.Context()).
		Where("is_active = ?", 1).
		Order("sort_order asc").
		Find(&dbareas).Error; err != nil {
		return errors.New("获取街道列表失败: " + err.Error())
	}

	areaMap := make(map[int][]types.Area)

	for _, a := range dbareas {
		areaMap[a.DistrictID] = append(areaMap[a.DistrictID], types.Area{
			ID:         a.ID,
			DistrictID: a.DistrictID,
			Name:       a.Name,
			SortOrder:  a.SortedOrder,
			IsActive:   a.IsActivate == 1,
		})
	}

	tree := make([]types.DistrictTree, 0, len(types.DistrictList))
	for _, d := range districts {
		areas := areaMap[d.ID]
		if areas == nil {
			areas = []types.Area{} // 保证 JSON 输出 [] 而不是 null
		}
		tree = append(tree, types.DistrictTree{
			ID:        d.ID,
			Name:      d.Name,
			SortOrder: d.SortedID,
			Areas:     areas,
		})
	}
	response.Success(c, tree)

	return nil
}

func formatMarkerTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

func queryInt(c *gin.Context, key string) int {
	value, _ := strconv.Atoi(c.Query(key))
	return value
}

func parseTagBits(raw string) int {
	var bits int
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		id, _ := strconv.Atoi(item)
		if id > 0 {
			bits |= id
		}
	}
	return bits
}

func tagBitsToIDs(bits int) []int {
	ids := make([]int, 0)
	for bit := 1; bit <= bits; bit <<= 1 {
		if bits&bit == bit {
			ids = append(ids, bit)
		}
	}
	return ids
}
