package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/models"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	stdcontext "context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Merchant struct {
	Config          *config.Config
	DB              *gorm.DB
	UserService     service.IUserService
	NoteService     service.INoteService
	MerchantService service.IMerchantService
	FollowService   service.IFollowService
}

func (pc *Merchant) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(pc.Config.Jwt.Secret))
	optionalAuth := middleware.OptionalAuth([]byte(pc.Config.Jwt.Secret))
	m := r.Group("/v1/merchant")
	{
		m.GET("/list", optionalAuth, context.Wrap(pc.GetPartyList))
		m.GET("/:id/follower/count", optionalAuth, context.Wrap(pc.GetPartyFollowerCount))
		m.GET("/:id/goods", optionalAuth, context.Wrap(pc.GetPartyGoods))
		m.GET("/:id", optionalAuth, context.Wrap(pc.GetPartyDetail))
		m.GET("/tags", optionalAuth, pc.GetTags)

		m.POST("/create", authorize, pc.CreateMerchant) // 创建商家

		m.POST("/:id/attend", authorize, pc.AttendParty)
		m.DELETE("/:id/attend", authorize, pc.CancelAttend)
		m.POST("/subscribe", authorize, context.Wrap(pc.subscribeParty))
		m.POST("/unsubscribe", authorize, context.Wrap(pc.UnsubcribParty))

	}
	c := r.Group("/v1/category")
	{
		c.GET("/list", optionalAuth, context.Wrap(pc.GetCategoryList))
	}
	s := r.Group("/v1/subscribe")
	{
		s.GET("/list", authorize, context.Wrap(pc.GetSubscribeList))
	}
}

func (pc *Merchant) GetPartyFollowerCount(c *gin.Context) error {
	partyID, err := strconv.Atoi(c.Param("id"))
	if err != nil || partyID <= 0 {
		return response.NewError(http.StatusBadRequest, "商家ID无效")
	}
	var party models.Merchant
	if err := pc.DB.WithContext(c.Request.Context()).First(&party, partyID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewError(http.StatusNotFound, "商家不存在")
		}
		return err
	}
	count, err := pc.FollowService.GetFollowerCount(c.Request.Context(), uint64(party.UserID))
	if err != nil {
		return err
	}
	response.Success(c, gin.H{"follower_count": count})
	return nil
}

func (pc *Merchant) GetPartyGoods(c *gin.Context) error {
	partyID, err := strconv.Atoi(c.Param("id"))
	if err != nil || partyID <= 0 {
		return response.NewError(http.StatusBadRequest, "商家ID无效")
	}
	goods := make([]models.Product, 0)
	if err := pc.DB.WithContext(c.Request.Context()).Where("party_id = ?", partyID).Find(&goods).Error; err != nil {
		return err
	}
	response.Success(c, gin.H{"list": goods, "total": len(goods)})
	return nil
}

func (pc *Merchant) GetSubscribeList(c *gin.Context) error {
	resp, err := pc.MerchantService.GetUserSubscribedParties(c, c.GetInt("user_id"))
	if err != nil {
		return response.NewError(500, err.Error())
	}
	response.Success(c, resp)
	return nil
}

func (pc *Merchant) GetCategoryList(c *gin.Context) error {
	var resp []models.Category
	err := pc.DB.
		Model(&models.Category{}).Find(&resp).Error
	if err != nil {
		return err
	}
	response.Success(c, resp)
	return nil
}

func (pc *Merchant) GetTags(c *gin.Context) {
	configs := []types.TagConfigResp{
		{Name: "积分立减", Id: 1},
		{Name: "买单立减", Id: 2},
		{Name: "新人优惠", Id: 4},
	}
	c.JSON(200, gin.H{"code": 200, "msg": "ok", "data": configs})

}

func (pc *Merchant) CreateMerchant(c *gin.Context) {
	var req types.CreatePartyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数格式错误"})
		return
	}

	userID := c.GetInt("user_id")

	// 转换为数据库模型
	party := models.Merchant{
		UserID:       userID,
		Title:        req.Title,
		Type:         req.Type,
		Description:  req.Description,
		LocationName: req.LocationName,
		Address:      req.Address,
		Latitude:     req.Lat,
		Longitude:    req.Lng,
		//Price:        int64(req.Price * 100), // 元转分
	}

	// 处理 JSON 图片字段
	//imagesByte, _ := json.Marshal(req.Images)
	//party.ImagesJSON = string(imagesByte)

	// 执行事务写入
	err := pc.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 插入主表
		// 注意：geo_point 在 MySQL 中是虚拟生成列，不需要手动插入
		if err := tx.Create(&party).Error; err != nil {
			return err
		}

		// 2. 批量插入标签
		if len(req.Tags) > 0 {
			var tags []models.PartyTag
			for _, t := range req.Tags {
				tags = append(tags, models.PartyTag{
					PartyID: party.ID,
					TagName: t,
				})
			}
			if err := tx.Create(&tags).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "发布失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"code": 200, "msg": "发布成功", "data": party.ID})
}

func (pc *Merchant) GetPartyLikeStatus(userID int, partyIDs []int) (map[int]bool, error) {
	result := make(map[int]bool, len(partyIDs))
	for _, pid := range partyIDs {
		result[pid] = false
	}
	if len(partyIDs) == 0 {
		return result, nil
	}

	var likedPartyIDs []int
	err := pc.DB.
		Model(&models.PartyLike{}).
		Select("party_id").
		Where("user_id = ? AND party_id IN ?", userID, partyIDs).
		Pluck("party_id", &likedPartyIDs).Error
	if err != nil {
		return nil, err
	}

	for _, pid := range likedPartyIDs {
		result[pid] = true
	}
	return result, nil
}
func (pc *Merchant) GetPartyList(c *gin.Context) error {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	ctx := c.Request.Context()
	query := pc.DB.Model(&models.Merchant{}).Where("status = ?", "active")
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("title LIKE ? OR location_name LIKE ? OR address LIKE ? OR description LIKE ?", like, like, like, like)
	}
	districtId := c.Query("district_id")
	if districtId == "" {
		districtId = c.Query("district")
	}
	districtIdNum := pc.resolveDistrictID(c, districtId)
	if districtIdNum > 0 {
		query = query.Where("district_id = ?", districtIdNum)
	}
	areaRaw := c.Query("area_id")
	if areaRaw == "" {
		areaRaw = c.Query("area")
	}
	areaID := pc.resolveAreaID(c, areaRaw)
	if areaID > 0 {
		query = query.Where("area_id = ?", areaID)
	}
	if businessArea := strings.TrimSpace(c.Query("business_area")); businessArea != "" {
		like := "%" + businessArea + "%"
		query = query.Where("location_name LIKE ? OR address LIKE ?", like, like)
	}
	categoryParam := c.Query("category")
	if categoryParam != "" {
		categoryStrings := strings.Split(categoryParam, ",")
		var filtered []string
		for _, v := range categoryStrings {
			v = strings.TrimSpace(v)
			if v != "" {
				filtered = append(filtered, v)
			}
		}
		if len(filtered) > 0 {
			query = query.Where("category IN ?", filtered)
		}
	}
	tagsParam := c.Query("tag_ids")
	if tagsParam == "" {
		tagsParam = c.Query("tags")
	}
	tagStrings := strings.Split(tagsParam, ",")
	var requiredTags int
	for _, s := range tagStrings {
		val, _ := strconv.Atoi(s)
		requiredTags |= val
	}
	if requiredTags > 0 {
		query = query.Where("tags & ? = ?", requiredTags, requiredTags)
	}
	lat, latErr := strconv.ParseFloat(c.Query("lat"), 64)
	lng, lngErr := strconv.ParseFloat(c.Query("lng"), 64)
	distanceLimit, _ := strconv.ParseFloat(c.Query("distance"), 64)
	if latErr == nil && lngErr == nil && lat != 0 && lng != 0 && distanceLimit > 0 {
		distanceSQL := `
        (6371 * acos(
            cos(radians(?)) *
            cos(radians(latitude)) *
            cos(radians(longitude) - radians(?)) +
            sin(radians(?)) *
            sin(radians(latitude))
        ))
        `
		query = query.Where(distanceSQL+" <= ?", lat, lng, lat, distanceLimit)
	}

	sortType := c.DefaultQuery("sort", "default")

	switch sortType {

	case "rating":
		query = query.Order("price DESC")

	case "popularity":
		// 推荐按 join_count + view_count 组合
		query = query.Order("district_id DESC")

	case "distance":
		if latErr == nil && lngErr == nil && lat != 0 && lng != 0 {
			// Haversine 公式计算距离（单位：km）
			distanceSQL := `
        (6371 * acos(
            cos(radians(?)) *
            cos(radians(latitude)) *
            cos(radians(longitude) - radians(?)) +
            sin(radians(?)) *
            sin(radians(latitude))
        ))
        `
			query = query.
				Select("*,"+distanceSQL+" AS distance", lat, lng, lat).
				Order("distance ASC")
		}

	default:
		query = query.Order("created_at DESC")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "查询失败", "data": nil})
		return nil
	}
	offset := (page - 1) * pageSize
	var merchant []models.Merchant

	if err := query.Offset(offset).Limit(pageSize).Find(&merchant).Error; err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "查询失败", "data": nil})
		return nil
	}

	userIDArr := make([]uint64, 0)
	memIdSlice := make([]int, 0)
	for _, m := range merchant {
		userIDArr = append(userIDArr, uint64(m.UserID))
		memIdSlice = append(memIdSlice, int(m.ID))
	}

	isSubcribe, _ := pc.GetPartyLikeStatus(c.GetInt("user_id"), memIdSlice)

	userMap := pc.UserService.BatchGetUserInfo(ctx, userIDArr)

	list := make([]models.MerchantListItem, 0, len(merchant))
	for _, m := range merchant {
		isFollow, _ := pc.FollowService.CheckFollowStatus(c, uint64(c.GetInt("user_id")), uint64(m.UserID))
		userId := uint64(m.UserID)
		userAvatar := userMap[userId].Avatar
		userName := userMap[userId].Nickname
		icon := ""
		if m.Type == "场地" {
			icon = "https://cdn.hypercn.cn/icon/jiuba.png"
		} else {
			icon = "https://cdn.hypercn.cn/icon/party.png"
		}
		list = append(list, models.MerchantListItem{
			ID:           m.ID,
			UserID:       m.UserID,
			UserAvatar:   userAvatar,
			UserName:     userName,
			CoverImage:   m.CoverImage,
			Title:        m.Title,
			Type:         m.Type,
			CreatedAt:    m.CreatedAt,
			Lat:          m.Latitude,
			Lng:          m.Longitude,
			Location:     m.LocationName,
			CurrentCount: 9932,
			AvgPrice:     7600,
			PostCount:    372,
			Icon:         icon,
			IsFollow:     isFollow,
			IsSubscriber: isSubcribe[int(m.ID)],
			CategoryID:   m.Category,
			DistrictID:   m.DistrictID,
			AreaID:       m.AreaID,
			TagIDs:       tagBitsToIDs(m.Tags),
		})
	}

	resp := types.PartyList{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
	response.Success(c, resp)
	return nil
}

func (pc *Merchant) resolveDistrictID(c *gin.Context, raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	if id, err := strconv.Atoi(raw); err == nil && id > 0 {
		return id
	}
	var district models.District
	if err := pc.DB.WithContext(c.Request.Context()).Where("name = ?", raw).First(&district).Error; err != nil {
		return 0
	}
	return district.ID
}

func (pc *Merchant) resolveAreaID(c *gin.Context, raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	if id, err := strconv.Atoi(raw); err == nil && id > 0 {
		return id
	}
	var area models.Areas
	if err := pc.DB.WithContext(c.Request.Context()).Where("name = ?", raw).First(&area).Error; err != nil {
		return 0
	}
	return area.ID
}

// GetPartyDetail 获取派对详情
// GET /api/v1/party/:id
func (pc *Merchant) GetPartyDetail(c *gin.Context) error {
	merchantID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || merchantID <= 0 {
		return response.NewError(http.StatusBadRequest, "商家ID无效")
	}

	var merchant models.Merchant
	err = pc.DB.WithContext(c.Request.Context()).Where("id = ?", merchantID).First(&merchant).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 首页已迁移到 organizers/activities。保留旧 merchant 详情路径，
		// 使仍使用该路径的客户端能够拿到新场地的 owner user_id。
		resp, fallbackErr := pc.getOrganizerMerchantDetail(c, merchantID)
		if fallbackErr != nil {
			if errors.Is(fallbackErr, gorm.ErrRecordNotFound) {
				return response.NewError(http.StatusNotFound, "商家不存在")
			}
			return fallbackErr
		}
		response.Success(c, resp)
		return nil
	}
	if err != nil {
		return err
	}

	resp := types.MerchantDetail{
		Id:            merchant.ID,
		UserId:        merchant.UserID,
		Name:          merchant.Title,
		AvgPrice:      7600,
		LocationName:  merchant.LocationName,
		Images:        make([]string, 0),
		Goods:         make([]models.Product, 0),
		BusinessHours: "19:30-次日02:30",
	}
	_ = json.Unmarshal([]byte(merchant.ImagesJSON), &resp.Images)
	if err := pc.DB.WithContext(c.Request.Context()).Where("party_id = ?", merchantID).Find(&resp.Goods).Error; err != nil {
		return err
	}
	resp.UserAvatar, resp.UserName, _ = pc.UserService.GetUserAvatar(c.Request.Context(), int64(merchant.UserID))
	if userID := c.GetInt("user_id"); userID > 0 {
		resp.IsFollow, _ = pc.FollowService.CheckFollowStatus(c, uint64(userID), uint64(merchant.UserID))
		isSubscribed, err := pc.MerchantService.CheckSubcribe(c, userID, int(merchant.ID))
		if err != nil {
			return response.NewError(http.StatusInternalServerError, "查询订阅状态失败")
		}
		resp.IsSubscribe = isSubscribed
	}

	response.Success(c, resp)
	return nil
}

// getOrganizerMerchantDetail is a backward-compatible detail adapter for the
// retired merchant endpoint. New venues belong to organizers, not parties.
func (pc *Merchant) getOrganizerMerchantDetail(c *gin.Context, organizerID int64) (*types.MerchantDetail, error) {
	var organizer models.Organizer
	if err := pc.DB.WithContext(c.Request.Context()).
		Where("id = ? AND status = ? AND enabled = 1", organizerID, models.OrganizerStatusApproved).
		First(&organizer).Error; err != nil {
		return nil, err
	}

	var profile models.OrganizerProfile
	if err := pc.DB.WithContext(c.Request.Context()).Where("organizer_id = ?", organizer.ID).First(&profile).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	images := make([]string, 0)
	if profile.Gallery != "" {
		_ = json.Unmarshal([]byte(profile.Gallery), &images)
	}
	if len(images) == 0 && profile.CoverImage != "" {
		images = append(images, profile.CoverImage)
	}
	resp := &types.MerchantDetail{
		Id:            organizer.ID,
		UserId:        int(organizer.UserID),
		Name:          organizer.Name,
		AvgPrice:      profile.AverageSpend,
		LocationName:  profile.Address,
		Images:        images,
		Goods:         make([]models.Product, 0),
		BusinessHours: profile.BusinessHours,
	}
	resp.UserAvatar, resp.UserName, _ = pc.UserService.GetUserAvatar(c.Request.Context(), organizer.UserID)
	if userID := c.GetInt("user_id"); userID > 0 {
		resp.IsFollow, _ = pc.FollowService.CheckFollowStatus(c, uint64(userID), uint64(organizer.UserID))
		var count int64
		if err := pc.DB.WithContext(c.Request.Context()).Model(&models.VenueSubscription{}).
			Where("organizer_id = ? AND user_id = ?", organizer.ID, userID).Count(&count).Error; err != nil {
			return nil, err
		}
		resp.IsSubscribe = count > 0
	}
	return resp, nil
}

// AttendParty 报名参加派对
// POST /api/v1/party/:id/attend
func (pc *Merchant) AttendParty(c *gin.Context) {
	partyID := c.Param("id")
	userID := c.GetInt("user_id")

	if userID == 0 {
		c.JSON(401, gin.H{"code": 401, "msg": "请先登录", "data": nil})
		return
	}

	// 检查派对是否存在
	var party models.Merchant
	if err := pc.DB.Where("id = ? AND deleted_at IS NULL", partyID).
		First(&party).Error; err != nil {
		c.JSON(404, gin.H{"code": 404, "msg": "派对不存在", "data": nil})
		return
	}
	//
	//// 检查派对状态
	//if party.Status == 3 {
	//	c.JSON(400, gin.H{"code": 400, "msg": "派对已结束", "data": nil})
	//	return
	//}
	//if party.Status == 4 {
	//	c.JSON(400, gin.H{"code": 400, "msg": "派对已取消", "data": nil})
	//	return
	//}

	// 检查是否已报名
	var count int64
	pc.DB.Model(&models.PartyAttendee{}).
		Where("party_id = ? AND user_id = ?", partyID, userID).
		Count(&count)

	if count > 0 {
		c.JSON(400, gin.H{"code": 400, "msg": "已经报名过了", "data": nil})
		return
	}

	// 创建报名记录
	attendee := models.PartyAttendee{
		PartyID: party.ID,
		UserID:  userID,
		Status:  1,
	}

	if err := pc.DB.Create(&attendee).Error; err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "报名失败", "data": nil})
		return
	}

	// 增加参与人数
	pc.DB.Model(&models.Merchant{}).Where("id = ?", partyID).
		UpdateColumn("attendee_count", gorm.Expr("attendee_count + 1"))

	// 更新热度分数
	pc.updateHotScore(party.ID)

	c.JSON(200, gin.H{"code": 200, "msg": "报名成功", "data": nil})
}

// CancelAttend 取消报名
// DELETE /api/v1/party/:id/attend
func (pc *Merchant) CancelAttend(c *gin.Context) {
	partyID := c.Param("id")
	userID := c.GetInt("user_id")

	if userID == 0 {
		c.JSON(401, gin.H{"code": 401, "msg": "请先登录", "data": nil})
		return
	}

	// 删除报名记录
	result := pc.DB.Where("party_id = ? AND user_id = ?", partyID, userID).
		Delete(&models.PartyAttendee{})

	if result.RowsAffected == 0 {
		c.JSON(400, gin.H{"code": 400, "msg": "未报名该派对", "data": nil})
		return
	}

	// 减少参与人数
	pc.DB.Model(&models.Merchant{}).Where("id = ?", partyID).
		UpdateColumn("attendee_count", gorm.Expr("attendee_count - 1"))

	// 更新热度分数
	partyIDUint, _ := strconv.ParseInt(partyID, 10, 64)
	pc.updateHotScore(partyIDUint)

	c.JSON(200, gin.H{"code": 200, "msg": "取消成功", "data": nil})
}

// 更新热度分数
func (pc *Merchant) updateHotScore(partyID int64) {
	var party models.Merchant
	if err := pc.DB.Where("id = ?", partyID).First(&party).Error; err != nil {
		return
	}
}

func (pc *Merchant) subscribeParty(c *gin.Context) error {
	var req types.SubcribPartyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}
	userId := c.GetInt("user_id")
	if pc.isActivityID(c.Request.Context(), req.PartyId) {
		sub := models.ActivitySubscription{ActivityID: req.PartyId, UserID: int64(userId)}
		if err := pc.DB.WithContext(c.Request.Context()).Clauses(clause.OnConflict{DoNothing: true}).Create(&sub).Error; err != nil {
			return response.NewError(http.StatusInternalServerError, "订阅失败")
		}
		response.Success(c, "订阅成功")
		return nil
	}
	err := pc.MerchantService.SubcribParty(c.Request.Context(), userId, int(req.PartyId))
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, "订阅成功")
	return nil
}

func (pc *Merchant) UnsubcribParty(c *gin.Context) error {
	var req types.UnsubcribPartyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "参数错误")
	}
	var userId int
	authHeader := c.GetHeader("Authorization")
	if authHeader == "Bearer debug-mode" {
		userId = 6 // Debug 模式下使用固定用户ID
	} else {
		userId = c.GetInt("user_id")
	}
	if pc.isActivityID(c.Request.Context(), req.PartyId) {
		if err := pc.DB.WithContext(c.Request.Context()).
			Where("activity_id = ? AND user_id = ?", req.PartyId, userId).
			Delete(&models.ActivitySubscription{}).Error; err != nil {
			return response.NewError(http.StatusInternalServerError, "取消订阅失败")
		}
		response.Success(c, "取消订阅成功")
		return nil
	}
	err := pc.MerchantService.UnsubcribParty(c.Request.Context(), int(userId), int(req.PartyId))
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, "取消订阅成功")
	return nil
}

func (pc *Merchant) isActivityID(ctx stdcontext.Context, id int64) bool {
	var count int64
	_ = pc.DB.WithContext(ctx).Model(&models.Activity{}).Where("id = ?", id).Count(&count).Error
	return count > 0
}
