package handler

import (
	"Hyper/config"
	"Hyper/dao"
	"Hyper/middleware"
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
	Config     *config.Config
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
	markerGroup.GET("/markers", middleware.OptionalAuth([]byte(m.Config.Jwt.Secret)), context.Wrap(m.GetMarkers))
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
	tagIDs, err := dao.ParseContentTagIDs(firstNonEmptyQuery(c, "tag_ids", "tags"))
	if err != nil {
		return response.NewError(400, err.Error())
	}

	markers := make([]types.MapMarker, 0)
	if source == "merchant" {
		response.Success(c, types.MapMarkerResponse{List: markers, Total: 0})
		return nil
	}
	if source == "all" || source == "activity" || source == "party" || source == "venue" {
		activities, err := m.getActivityMarkers(c, limit, tagIDs)
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

func (m *Map) getPartyMarkers(c *gin.Context, limit int, tagIDs []int64) ([]types.MapMarker, error) {
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
	query = dao.ApplyContentTagFilter(query, models.ContentTagTargetParty, "parties.id", tagIDs)
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
	partyIDs := make([]int64, 0, len(parties))
	for _, party := range parties {
		partyIDs = append(partyIDs, party.ID)
	}
	tagMap, err := dao.LoadContentTags(c.Request.Context(), m.DB, models.ContentTagTargetParty, partyIDs, false)
	if err != nil {
		return nil, err
	}
	followCounts, followed, err := dao.LoadContentFollowStats(c.Request.Context(), m.DB, models.ContentFollowTargetParty, partyIDs, int64(currentUserID(c)))
	if err != nil {
		return nil, err
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
		tags := tagMap[party.ID]
		marker := types.MapMarker{
			ID:               fmt.Sprintf("party-%d", party.ID),
			Source:           source,
			SourceID:         party.ID,
			DetailType:       "merchant",
			DetailURL:        fmt.Sprintf("/api/v1/merchant/%d", party.ID),
			UserID:           int64(party.UserID),
			Title:            party.Title,
			Type:             party.Type,
			Location:         party.LocationName,
			Address:          party.Address,
			Lat:              party.Latitude,
			Lng:              party.Longitude,
			CoverImage:       party.CoverImage,
			CreatedAt:        formatMarkerTime(party.CreatedAt),
			AvgPrice:         7600,
			CurrentCount:     9932,
			PostCount:        372,
			Icon:             icon,
			Status:           party.Status,
			CategoryID:       party.Category,
			DistrictID:       party.DistrictID,
			AreaID:           party.AreaID,
			TagIDs:           types.ContentTagIDs(tags),
			DiscountTags:     types.ContentTagNames(tags),
			Tags:             types.BuildContentTagItems(tags),
			IsFollow:         followed[party.ID],
			FollowCount:      followCounts[party.ID],
			FollowTargetType: models.ContentFollowTargetParty,
			FollowTargetID:   party.ID,
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

func (m *Map) getActivityMarkers(c *gin.Context, limit int, tagIDs []int64) ([]types.MapMarker, error) {
	var activities []models.Activity
	query := m.DB.WithContext(c.Request.Context()).
		Table("activities AS a").
		Select("a.*").
		Joins("JOIN organizers o ON o.id = a.organizer_id").
		Where("a.status = ? AND a.is_hidden = 0 AND a.latitude <> 0 AND a.longitude <> 0", models.ActivityStatusOnline).
		Where("o.status = ? AND o.enabled = 1", models.OrganizerStatusApproved).
		Where("(a.type = ? OR a.end_time >= ?)", models.ActivityTypeVenue, time.Now()).
		Where("(a.type <> ? OR a.id = (SELECT MAX(av.id) FROM activities av WHERE av.organizer_id = a.organizer_id AND av.type = ? AND av.status = ? AND av.is_hidden = 0))", models.ActivityTypeVenue, models.ActivityTypeVenue, models.ActivityStatusOnline)
	if activityType := m.resolveActivityTypeFilter(c); activityType != "" {
		query = query.Where("a.type = ?", activityType)
	}
	if len(tagIDs) > 0 {
		activityMatch, activityArgs := dao.ContentTagMatchSQL(models.ContentTagTargetActivity, "a.id", tagIDs)
		venueMatch, venueArgs := dao.ContentTagMatchSQL(models.ContentTagTargetVenue, "a.organizer_id", tagIDs)
		args := make([]any, 0, len(activityArgs)+len(venueArgs)+2)
		args = append(args, models.ActivityTypeVenue)
		args = append(args, venueArgs...)
		args = append(args, models.ActivityTypeVenue)
		args = append(args, activityArgs...)
		query = query.Where("((a.type = ? AND "+venueMatch+") OR (a.type <> ? AND "+activityMatch+"))", args...)
	}
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("a.name LIKE ? OR a.address LIKE ? OR a.province LIKE ? OR a.city LIKE ? OR a.district LIKE ? OR a.description LIKE ? OR o.name LIKE ?", like, like, like, like, like, like, like)
	}
	if district := strings.TrimSpace(c.Query("district")); district != "" {
		query = query.Where("a.district = ?", district)
	}
	if districtName := m.resolveDistrictName(c, c.Query("district_id")); districtName != "" {
		query = query.Where("a.district = ?", districtName)
	}
	if area := strings.TrimSpace(c.Query("area")); area != "" {
		query = query.Where("a.address LIKE ?", "%"+area+"%")
	}
	if businessArea := strings.TrimSpace(c.Query("business_area")); businessArea != "" {
		query = query.Where("a.address LIKE ?", "%"+businessArea+"%")
	}
	if err := query.Order("a.created_at DESC").Limit(limit).Scan(&activities).Error; err != nil {
		return nil, err
	}

	organizerIDs := make([]int64, 0, len(activities))
	activityIDs := make([]int64, 0, len(activities))
	activityFollowIDs := make([]int64, 0, len(activities))
	venueFollowIDs := make([]int64, 0, len(activities))
	for _, activity := range activities {
		organizerIDs = append(organizerIDs, activity.OrganizerID)
		activityIDs = append(activityIDs, activity.ID)
		if activity.Type == models.ActivityTypeVenue {
			venueFollowIDs = append(venueFollowIDs, activity.OrganizerID)
		} else {
			activityFollowIDs = append(activityFollowIDs, activity.ID)
		}
	}
	activityTags, err := dao.LoadContentTags(c.Request.Context(), m.DB, models.ContentTagTargetActivity, activityIDs, false)
	if err != nil {
		return nil, err
	}
	venueTags, err := dao.LoadContentTags(c.Request.Context(), m.DB, models.ContentTagTargetVenue, organizerIDs, false)
	if err != nil {
		return nil, err
	}

	organizerMap := make(map[int64]models.Organizer)
	if len(organizerIDs) > 0 {
		var organizers []models.Organizer
		if err := m.DB.WithContext(c.Request.Context()).Where("id IN ?", organizerIDs).Find(&organizers).Error; err != nil {
			return nil, err
		}
		for _, organizer := range organizers {
			organizerMap[organizer.ID] = organizer
		}
	}

	subscribedMap := m.loadActivitySubscriptionSet(c, activityIDs)
	activityFollowCounts, activityFollowed, err := dao.LoadContentFollowStats(c.Request.Context(), m.DB, models.ContentFollowTargetActivity, activityFollowIDs, int64(currentUserID(c)))
	if err != nil {
		return nil, err
	}
	venueFollowCounts, venueFollowed, err := dao.LoadContentFollowStats(c.Request.Context(), m.DB, models.ContentFollowTargetVenue, venueFollowIDs, int64(currentUserID(c)))
	if err != nil {
		return nil, err
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
		tags := activityTags[activity.ID]
		followCount := activityFollowCounts[activity.ID]
		isFollow := activityFollowed[activity.ID]
		followTargetType := models.ContentFollowTargetActivity
		followTargetID := activity.ID
		if activityType == models.ActivityTypeVenue {
			tags = venueTags[activity.OrganizerID]
			followCount = venueFollowCounts[activity.OrganizerID]
			isFollow = venueFollowed[activity.OrganizerID]
			followTargetType = models.ContentFollowTargetVenue
			followTargetID = activity.OrganizerID
		}
		markerID := fmt.Sprintf("activity-%d", activity.ID)
		markerSource := "activity"
		markerSourceID := activity.ID
		detailType := "activity"
		detailURL := fmt.Sprintf("/api/v1/activity/%d", activity.ID)
		if activityType == models.ActivityTypeVenue {
			markerID = fmt.Sprintf("venue-%d", activity.OrganizerID)
			markerSource = "venue"
			markerSourceID = activity.OrganizerID
			detailType = "venue"
			detailURL = fmt.Sprintf("/api/v1/venues/%d", activity.OrganizerID)
		}
		marker := types.MapMarker{
			ID:               markerID,
			Source:           markerSource,
			SourceID:         markerSourceID,
			DetailType:       detailType,
			DetailURL:        detailURL,
			Title:            activity.Name,
			Type:             activityType,
			Location:         activity.Address,
			Address:          activity.Address,
			Lat:              activity.Latitude,
			Lng:              activity.Longitude,
			CoverImage:       activity.PosterList,
			CreatedAt:        formatMarkerTime(activity.CreatedAt),
			AvgPrice:         priceMap[activity.ID],
			CurrentCount:     0,
			PostCount:        0,
			Icon:             icon,
			IsFollow:         isFollow,
			FollowCount:      followCount,
			FollowTargetType: followTargetType,
			FollowTargetID:   followTargetID,
			StartTime:        formatMarkerTime(activity.StartTime),
			EndTime:          formatMarkerTime(activity.EndTime),
			Status:           activity.Status,
			TagIDs:           types.ContentTagIDs(tags),
			District:         activity.District,
			Distance:         markerDistance(c, activity.Latitude, activity.Longitude),
			SupportPoints:    true,
			DiscountTags:     types.ContentTagNames(tags),
			Tags:             types.BuildContentTagItems(tags),
		}
		if exceedsDistance(c, marker.Distance) {
			continue
		}
		if organizer, ok := organizerMap[activity.OrganizerID]; ok {
			if activityType == models.ActivityTypeVenue && organizer.Name != "" {
				marker.Title = organizer.Name
			}
			marker.UserID = organizer.UserID
			marker.User = organizer.Name
			marker.UserName = organizer.Name
			marker.UserAvatar = organizer.Logo
			marker.UserAvatarCamel = organizer.Logo
			marker.IsSubscriber = subscribedMap[activity.ID]
		}
		markers = append(markers, marker)
	}
	return markers, nil
}

// resolveActivityTypeFilter keeps category switching compatible while activity
// type is stored directly on activities rather than in the legacy categories table.
// Explicit source/type takes precedence over category_id.
func (m *Map) resolveActivityTypeFilter(c *gin.Context) string {
	switch strings.ToLower(strings.TrimSpace(c.Query("source"))) {
	case models.ActivityTypeParty:
		return models.ActivityTypeParty
	case models.ActivityTypeVenue:
		return models.ActivityTypeVenue
	}
	switch strings.ToLower(strings.TrimSpace(c.Query("type"))) {
	case models.ActivityTypeParty:
		return models.ActivityTypeParty
	case models.ActivityTypeVenue:
		return models.ActivityTypeVenue
	}

	categoryID := queryInt(c, "category_id")
	if categoryID <= 0 {
		return ""
	}

	// Prefer the configured category name so a deployment can customize IDs.
	var category models.Category
	if err := m.DB.WithContext(c.Request.Context()).First(&category, categoryID).Error; err == nil {
		name := strings.ToLower(strings.TrimSpace(category.Name))
		switch {
		case strings.Contains(name, "场地"), strings.Contains(name, "venue"):
			return models.ActivityTypeVenue
		case strings.Contains(name, "派对"), strings.Contains(name, "party"):
			return models.ActivityTypeParty
		}
	}

	// Legacy category data uses 1 for venues and 2 for parties. Keep this
	// fallback so existing mini-program category selectors work without changes.
	switch categoryID {
	case 1:
		return models.ActivityTypeVenue
	case 2:
		return models.ActivityTypeParty
	default:
		return ""
	}
}

// loadActivitySubscriptionSet batch-loads current user's subscriptions for
// activity markers. Anonymous requests intentionally return an empty set.
func (m *Map) loadActivitySubscriptionSet(c *gin.Context, activityIDs []int64) map[int64]bool {
	subscribed := make(map[int64]bool)
	userID := currentUserID(c)
	if userID <= 0 || len(activityIDs) == 0 {
		return subscribed
	}
	var subscribedIDs []int64
	if err := m.DB.WithContext(c.Request.Context()).Model(&models.ActivitySubscription{}).
		Where("user_id = ? AND activity_id IN ?", userID, activityIDs).
		Pluck("activity_id", &subscribedIDs).Error; err != nil {
		return subscribed
	}
	for _, id := range subscribedIDs {
		subscribed[id] = true
	}
	return subscribed
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

func firstNonEmptyQuery(c *gin.Context, keys ...string) string {
	for _, key := range keys {
		if value := c.Query(key); value != "" {
			return value
		}
	}
	return ""
}

func markerDistance(c *gin.Context, lat, lng float64) float64 {
	userLat, latErr := strconv.ParseFloat(c.Query("lat"), 64)
	userLng, lngErr := strconv.ParseFloat(c.Query("lng"), 64)
	if latErr != nil || lngErr != nil || userLat == 0 || userLng == 0 || lat == 0 || lng == 0 {
		return 0
	}
	return haversineKM(userLat, userLng, lat, lng)
}

// loadFollowingSet 批量查询当前登录用户已关注的目标用户 ID，避免每个 marker 单独查询
func (m *Map) loadFollowingSet(c *gin.Context, ownerUserIDs []int) map[int]bool {
	following := make(map[int]bool)
	userID := currentUserID(c)
	if userID <= 0 || len(ownerUserIDs) == 0 {
		return following
	}
	var followeeIDs []int
	if err := m.DB.WithContext(c.Request.Context()).Model(&models.UserFollow{}).
		Where("follower_id = ? AND followee_id IN ? AND status = 1", userID, ownerUserIDs).
		Pluck("followee_id", &followeeIDs).Error; err != nil {
		return following
	}
	for _, id := range followeeIDs {
		following[id] = true
	}
	return following
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
