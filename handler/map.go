package handler

import (
	"Hyper/models"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	"errors"
	"fmt"
	"math"
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
	if source != "all" && source != "activity" && source != "party" && source != "venue" && source != "merchant" {
		return errors.New("source 仅支持 all、activity、party、venue、merchant")
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "200"))
	if limit <= 0 {
		limit = 200
	}
	if limit > 500 {
		limit = 500
	}

	markers := make([]types.MapMarker, 0)
	if source == "merchant" {
		response.Success(c, types.MapMarkerResponse{List: markers, Total: 0})
		return nil
	}
	if source == "all" || source == "activity" || source == "party" || source == "venue" {
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
	if districtID := m.resolveDistrictID(c, c.Query("district")); districtID > 0 {
		query = query.Where("district_id = ?", districtID)
	}
	if areaID := queryInt(c, "area_id"); areaID > 0 {
		query = query.Where("area_id = ?", areaID)
	}
	if areaID := m.resolveAreaID(c, c.Query("area")); areaID > 0 {
		query = query.Where("area_id = ?", areaID)
	}
	if tagBits := parseTagBits(c.Query("tag_ids")); tagBits > 0 {
		query = query.Where("tags & ? = ?", tagBits, tagBits)
	}
	if tagBits := parseTagBits(c.Query("tags")); tagBits > 0 {
		query = query.Where("tags & ? = ?", tagBits, tagBits)
	}
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("title LIKE ? OR location_name LIKE ? OR address LIKE ? OR description LIKE ?", like, like, like, like)
	}
	if businessArea := strings.TrimSpace(c.Query("business_area")); businessArea != "" {
		like := "%" + businessArea + "%"
		query = query.Where("location_name LIKE ? OR address LIKE ?", like, like)
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
			DetailType:   "merchant",
			DetailURL:    fmt.Sprintf("/api/v1/merchant/%d", party.ID),
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
			DiscountTags: []string{},
		}
		marker.Distance = markerDistance(c, marker.Lat, marker.Lng)
		if exceedsDistance(c, marker.Distance) {
			continue
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
	switch c.DefaultQuery("source", "all") {
	case "party":
		query = query.Where("type = ?", models.ActivityTypeParty)
	case "venue":
		query = query.Where("type = ?", models.ActivityTypeVenue)
	}
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR address LIKE ? OR province LIKE ? OR city LIKE ? OR district LIKE ? OR description LIKE ?", like, like, like, like, like, like)
	}
	if district := strings.TrimSpace(c.Query("district")); district != "" {
		query = query.Where("district = ?", district)
	}
	if districtName := m.resolveDistrictName(c, c.Query("district_id")); districtName != "" {
		query = query.Where("district = ?", districtName)
	}
	if area := strings.TrimSpace(c.Query("area")); area != "" {
		query = query.Where("address LIKE ?", "%"+area+"%")
	}
	if businessArea := strings.TrimSpace(c.Query("business_area")); businessArea != "" {
		query = query.Where("address LIKE ?", "%"+businessArea+"%")
	}
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
		activityType := activity.Type
		if activityType != models.ActivityTypeVenue {
			activityType = models.ActivityTypeParty
		}
		icon := "https://cdn.hypercn.cn/icon/party.png"
		if activityType == models.ActivityTypeVenue {
			icon = "https://cdn.hypercn.cn/icon/jiuba.png"
		}
		marker := types.MapMarker{
			ID:            fmt.Sprintf("activity-%d", activity.ID),
			Source:        "activity",
			SourceID:      activity.ID,
			DetailType:    "activity",
			DetailURL:     fmt.Sprintf("/api/v1/activity/%d", activity.ID),
			Title:         activity.Name,
			Type:          activityType,
			Location:      activity.Address,
			Address:       activity.Address,
			Lat:           activity.Latitude,
			Lng:           activity.Longitude,
			CoverImage:    activity.PosterList,
			CreatedAt:     formatMarkerTime(activity.CreatedAt),
			AvgPrice:      priceMap[activity.ID],
			CurrentCount:  0,
			PostCount:     0,
			Icon:          icon,
			StartTime:     formatMarkerTime(activity.StartTime),
			EndTime:       formatMarkerTime(activity.EndTime),
			Status:        activity.Status,
			District:      activity.District,
			Distance:      markerDistance(c, activity.Latitude, activity.Longitude),
			SupportPoints: true,
			DiscountTags:  []string{},
		}
		if exceedsDistance(c, marker.Distance) {
			continue
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

func (m *Map) resolveDistrictID(c *gin.Context, raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	if id, err := strconv.Atoi(raw); err == nil && id > 0 {
		return id
	}
	var district models.District
	if err := m.DB.WithContext(c.Request.Context()).Where("name = ?", raw).First(&district).Error; err != nil {
		return 0
	}
	return district.ID
}

func (m *Map) resolveDistrictName(c *gin.Context, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if _, err := strconv.Atoi(raw); err != nil {
		return raw
	}
	var district models.District
	if err := m.DB.WithContext(c.Request.Context()).Where("id = ?", raw).First(&district).Error; err != nil {
		return ""
	}
	return district.Name
}

func (m *Map) resolveAreaID(c *gin.Context, raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	if id, err := strconv.Atoi(raw); err == nil && id > 0 {
		return id
	}
	var area models.Areas
	if err := m.DB.WithContext(c.Request.Context()).Where("name = ?", raw).First(&area).Error; err != nil {
		return 0
	}
	return area.ID
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

func markerDistance(c *gin.Context, lat, lng float64) float64 {
	userLat, latErr := strconv.ParseFloat(c.Query("lat"), 64)
	userLng, lngErr := strconv.ParseFloat(c.Query("lng"), 64)
	if latErr != nil || lngErr != nil || userLat == 0 || userLng == 0 || lat == 0 || lng == 0 {
		return 0
	}
	return haversineKM(userLat, userLng, lat, lng)
}

func exceedsDistance(c *gin.Context, distance float64) bool {
	limit, err := strconv.ParseFloat(c.Query("distance"), 64)
	if err != nil || limit <= 0 || distance <= 0 {
		return false
	}
	return distance > limit
}

func haversineKM(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusKM = 6371.0
	toRad := func(v float64) float64 { return v * math.Pi / 180 }
	dLat := toRad(lat2 - lat1)
	dLng := toRad(lng2 - lng1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*math.Sin(dLng/2)*math.Sin(dLng/2)
	return earthRadiusKM * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}
