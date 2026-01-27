package handler

import (
	"Hyper/models"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/types"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"Hyper/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Party struct {
	DB *gorm.DB
}

func (pc *Party) RegisterRouter(r gin.IRouter) {
	party := r.Group("/v1/party")
	{
		// 公开接口
		party.GET("/list", context.Wrap(pc.GetPartyList)) // 获取派对列表
		party.GET("/:id", pc.GetPartyDetail)              // 获取派对详情
		party.POST("/create", pc.CreateParty)             // 新增发布接口
		party.POST("/:id/attend", pc.AttendParty)         // 报名
		party.DELETE("/:id/attend", pc.CancelAttend)      // 取消报名
		party.POST("/:id/like", pc.LikeParty)             // 点赞
		party.DELETE("/:id/like", pc.UnlikeParty)         // 取消点赞
	}
}

func (pc *Party) CreateParty(c *gin.Context) {
	var req types.CreatePartyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数格式错误"})
		return
	}

	userID := c.GetInt("user_id")
	startTime, _ := time.Parse("2006-01-02 15:04:05", req.StartTime)

	// 转换为数据库模型
	party := models.Party{
		UserID:       userID,
		Title:        req.Title,
		Type:         req.Type,
		Description:  req.Description,
		LocationName: req.LocationName,
		Address:      req.Address,
		Latitude:     req.Lat,
		Longitude:    req.Lng,
		Price:        int64(req.Price * 100), // 元转分
		MaxAttendees: int64(req.MaxAttendees),
		StartTime:    startTime,
		CoverImage:   req.CoverImage,
		Status:       1, // 默认为预告状态
	}

	// 处理 JSON 图片字段
	imagesByte, _ := json.Marshal(req.Images)
	party.ImagesJSON = string(imagesByte)

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

func (pc *Party) GetPartyList(c *gin.Context) error {
	// 获取参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	// 筛选参数
	partyType := c.Query("type")
	status := c.Query("status")
	keyword := c.Query("keyword")

	// 位置参数
	userLat, _ := strconv.ParseFloat(c.Query("lat"), 64)
	userLng, _ := strconv.ParseFloat(c.Query("lng"), 64)
	maxDistance, _ := strconv.ParseFloat(c.DefaultQuery("maxDistance", "0"), 64) // km

	// 排序参数
	sortBy := c.DefaultQuery("sortBy", "hot") // hot-热度, time-时间, new-最新, distance-距离

	// 时间筛选
	startDate := c.Query("startDate") // 格式：2025-12-31
	endDate := c.Query("endDate")

	// 价格筛选
	minPrice, _ := strconv.Atoi(c.DefaultQuery("minPrice", "0"))
	maxPrice, _ := strconv.Atoi(c.DefaultQuery("maxPrice", "0"))

	// 获取当前用户ID（用于判断是否已点赞、已报名）
	userID := c.GetInt("user_id") // 从JWT中获取，未登录为0

	// 构建查询
	query := pc.DB.Model(&models.Party{}).Where("deleted_at IS NULL")

	// 类型筛选
	if partyType != "" {
		query = query.Where("type = ?", partyType)
	}

	// 状态筛选
	if status != "" {
		query = query.Where("status = ?", status)
	} else {
		// 默认只显示未结束的派对
		query = query.Where("status IN (1, 2)")
	}

	// 关键词搜索
	if keyword != "" {
		query = query.Where("title LIKE ? OR location LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%")
	}

	// 时间筛选
	if startDate != "" {
		query = query.Where("DATE(start_time) >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("DATE(start_time) <= ?", endDate)
	}

	// 价格筛选
	if minPrice > 0 {
		query = query.Where("price >= ?", minPrice)
	}
	if maxPrice > 0 {
		query = query.Where("price <= ?", maxPrice)
	}

	// 距离筛选（如果提供了位置和最大距离）
	var partyIDs []int64
	if maxDistance > 0 && userLat != 0 && userLng != 0 {
		// 先查询所有派对的位置
		var allParties []models.Party
		pc.DB.Model(&models.Party{}).
			Where("deleted_at IS NULL").
			Select("id, latitude, longitude").
			Find(&allParties)

		// 筛选距离范围内的派对
		for _, party := range allParties {
			distance := utils.CalculateDistance(userLat, userLng,
				party.Latitude, party.Longitude)
			if distance <= maxDistance {
				partyIDs = append(partyIDs, party.ID)
			}
		}

		if len(partyIDs) > 0 {
			query = query.Where("id IN ?", partyIDs)
		} else {
			// 没有符合距离条件的派对
			c.JSON(200, gin.H{
				"code": 200,
				"msg":  "ok",
				"data": gin.H{
					"list":       []models.PartyListItem{},
					"total":      0,
					"page":       page,
					"pageSize":   pageSize,
					"totalPages": 0,
				},
			})
			return nil
		}
	}

	// 排序
	switch sortBy {
	case "time":
		query = query.Order("start_time ASC")
	case "new":
		query = query.Order("created_at DESC")
	case "distance":
		// 距离排序需要在应用层处理
		query = query.Order("id DESC")
	case "hot":
		fallthrough
	default:
		query = query.Order("hot_score DESC, attendee_count DESC, view_count DESC")
	}

	// 计算总数
	var total int64
	countQuery := query.Session(&gorm.Session{})
	if err := countQuery.Count(&total).Error; err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "查询失败", "data": nil})
		return nil
	}

	// 分页
	offset := (page - 1) * pageSize
	var parties []models.Party

	// 预加载关联数据
	if err := query.Preload("Tags").Preload("User").
		Offset(offset).Limit(pageSize).Find(&parties).Error; err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "查询失败", "data": nil})
		return nil
	}

	// 如果需要判断用户是否已点赞、已报名
	var likedPartyIDs []int64
	var attendingPartyIDs []int64

	if userID > 0 && len(parties) > 0 {
		partyIDs := make([]int64, len(parties))
		for i, p := range parties {
			partyIDs[i] = p.ID
		}

		// 查询已点赞的派对
		var likes []models.PartyLike
		pc.DB.Where("user_id = ? AND party_id IN ?", userID, partyIDs).
			Find(&likes)
		for _, like := range likes {
			likedPartyIDs = append(likedPartyIDs, like.PartyID)
		}

		// 查询已报名的派对
		var attendees []models.PartyAttendee
		pc.DB.Where("user_id = ? AND party_id IN ? AND status = 1",
			userID, partyIDs).Find(&attendees)
		for _, attendee := range attendees {
			attendingPartyIDs = append(attendingPartyIDs, attendee.PartyID)
		}
	}

	// 转换为前端格式
	list := make([]models.PartyListItem, 0, len(parties))
	for _, party := range parties {
		item := convertToPartyListItem(party, userLat, userLng)
		item.DynamicCount = 55
		if item.ID%2 == 1 {
			item.Tag = "场地"
		} else {
			item.Tag = "派对"
		}

		// 设置是否已点赞
		item.IsLiked = utils.Contains(likedPartyIDs, party.ID)

		// 设置是否已报名
		item.IsAttending = utils.Contains(attendingPartyIDs, party.ID)

		list = append(list, item)

	}

	// 如果是按距离排序，在这里排序
	if sortBy == "distance" && userLat != 0 && userLng != 0 {
		sort.Slice(list, func(i, j int) bool {
			distI, _ := strconv.ParseFloat(
				strings.TrimSuffix(list[i].Distance, "km"), 64)
			distJ, _ := strconv.ParseFloat(
				strings.TrimSuffix(list[j].Distance, "km"), 64)
			return distI < distJ
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

// GetPartyDetail 获取派对详情
// GET /api/v1/party/:id
func (pc *Party) GetPartyDetail(c *gin.Context) {
	partyID := c.Param("id")
	userID := c.GetInt("user_id")

	var party models.Party
	if err := pc.DB.Preload("Tags").Preload("User").
		Where("id = ? AND deleted_at IS NULL", partyID).
		First(&party).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, gin.H{"code": 404, "msg": "派对不存在", "data": nil})
		} else {
			c.JSON(500, gin.H{"code": 500, "msg": "查询失败", "data": nil})
		}
		return
	}

	// 增加浏览次数
	go func() {
		pc.DB.Model(&models.Party{}).Where("id = ?", partyID).
			UpdateColumn("view_count", gorm.Expr("view_count + 1"))
	}()

	// 检查当前用户是否已报名
	isAttending := false
	if userID > 0 {
		var count int64
		pc.DB.Model(&models.PartyAttendee{}).
			Where("party_id = ? AND user_id = ? AND status = 1", partyID, userID).
			Count(&count)
		isAttending = count > 0
	}

	// 检查当前用户是否已点赞
	isLiked := false
	if userID > 0 {
		var count int64
		pc.DB.Model(&models.PartyLike{}).
			Where("party_id = ? AND user_id = ?", partyID, userID).
			Count(&count)
		isLiked = count > 0
	}

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": gin.H{
			"party":       party,
			"isAttending": isAttending,
			"isLiked":     isLiked,
		},
	})
}

// AttendParty 报名参加派对
// POST /api/v1/party/:id/attend
func (pc *Party) AttendParty(c *gin.Context) {
	partyID := c.Param("id")
	userID := c.GetInt("user_id")

	if userID == 0 {
		c.JSON(401, gin.H{"code": 401, "msg": "请先登录", "data": nil})
		return
	}

	// 检查派对是否存在
	var party models.Party
	if err := pc.DB.Where("id = ? AND deleted_at IS NULL", partyID).
		First(&party).Error; err != nil {
		c.JSON(404, gin.H{"code": 404, "msg": "派对不存在", "data": nil})
		return
	}

	// 检查派对状态
	if party.Status == 3 {
		c.JSON(400, gin.H{"code": 400, "msg": "派对已结束", "data": nil})
		return
	}
	if party.Status == 4 {
		c.JSON(400, gin.H{"code": 400, "msg": "派对已取消", "data": nil})
		return
	}

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
	pc.DB.Model(&models.Party{}).Where("id = ?", partyID).
		UpdateColumn("attendee_count", gorm.Expr("attendee_count + 1"))

	// 更新热度分数
	pc.updateHotScore(party.ID)

	c.JSON(200, gin.H{"code": 200, "msg": "报名成功", "data": nil})
}

// CancelAttend 取消报名
// DELETE /api/v1/party/:id/attend
func (pc *Party) CancelAttend(c *gin.Context) {
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
	pc.DB.Model(&models.Party{}).Where("id = ?", partyID).
		UpdateColumn("attendee_count", gorm.Expr("attendee_count - 1"))

	// 更新热度分数
	partyIDUint, _ := strconv.ParseInt(partyID, 10, 64)
	pc.updateHotScore(partyIDUint)

	c.JSON(200, gin.H{"code": 200, "msg": "取消成功", "data": nil})
}

// LikeParty 点赞派对
// POST /api/v1/party/:id/like
func (pc *Party) LikeParty(c *gin.Context) {
	partyID := c.Param("id")
	userID := c.GetInt("user_id")

	if userID == 0 {
		c.JSON(401, gin.H{"code": 401, "msg": "请先登录", "data": nil})
		return
	}

	partyIDUint, _ := strconv.ParseInt(partyID, 10, 64)

	// 检查是否已点赞
	var count int64
	pc.DB.Model(&models.PartyLike{}).
		Where("party_id = ? AND user_id = ?", partyID, userID).
		Count(&count)

	if count > 0 {
		c.JSON(400, gin.H{"code": 400, "msg": "已经点赞过了", "data": nil})
		return
	}

	// 创建点赞记录
	like := models.PartyLike{
		PartyID: partyIDUint,
		UserID:  userID,
	}

	if err := pc.DB.Create(&like).Error; err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "点赞失败", "data": nil})
		return
	}

	// 增加点赞数
	pc.DB.Model(&models.Party{}).Where("id = ?", partyID).
		UpdateColumn("like_count", gorm.Expr("like_count + 1"))

	// 更新热度分数
	pc.updateHotScore(partyIDUint)

	c.JSON(200, gin.H{"code": 200, "msg": "点赞成功", "data": nil})
}

// UnlikeParty 取消点赞
// DELETE /api/v1/party/:id/like
func (pc *Party) UnlikeParty(c *gin.Context) {
	partyID := c.Param("id")
	userID := c.GetInt("user_id")

	if userID == 0 {
		c.JSON(401, gin.H{"code": 401, "msg": "请先登录", "data": nil})
		return
	}

	// 删除点赞记录
	result := pc.DB.Where("party_id = ? AND user_id = ?", partyID, userID).
		Delete(&models.PartyLike{})

	if result.RowsAffected == 0 {
		c.JSON(400, gin.H{"code": 400, "msg": "未点赞该派对", "data": nil})
		return
	}

	// 减少点赞数
	pc.DB.Model(&models.Party{}).Where("id = ?", partyID).
		UpdateColumn("like_count", gorm.Expr("like_count - 1"))

	// 更新热度分数
	partyIDUint, _ := strconv.ParseInt(partyID, 10, 64)
	pc.updateHotScore(partyIDUint)

	c.JSON(200, gin.H{"code": 200, "msg": "取消成功", "data": nil})
}

func convertToPartyListItem(party models.Party, userLat, userLng float64) models.PartyListItem {
	// 1. 计算距离 (保持内存计算作为兜底，如果 SQL 没算的话)
	distance := ""
	if userLat != 0 && userLng != 0 {
		dist := utils.CalculateDistance(userLat, userLng, party.Latitude, party.Longitude)
		if dist < 1 {
			distance = fmt.Sprintf("%.0fm", dist*1000) // 小于1km显示米
		} else {
			distance = fmt.Sprintf("%.2fkm", dist)
		}
	}

	// 2. 提取标签 (TagName 匹配新 Model)
	tags := make([]string, 0, len(party.Tags))
	for _, tag := range party.Tags {
		tags = append(tags, tag.TagName)
	}

	// 3. 格式化时间
	timeStr := party.StartTime.Format("2006.01.02")

	// 4. 格式化价格 (精确保留小数)
	// 如果价格是 0，可以显示 "免费"
	priceStr := "免费"
	if party.Price > 0 {
		priceStr = fmt.Sprintf("%.2f", float64(party.Price)/100.0)
		// 如果是整数，去掉末尾的 .00
		priceStr = strings.TrimSuffix(priceStr, ".00")
	}

	userName := "未知用户"
	userAvatar := ""
	if party.User.Id > 0 { // 确保 Preload("User") 成功加载了数据
		userName = party.User.Nickname
		userAvatar = party.User.Avatar
	}

	// 6. 判断是否满员
	isFull := false
	if party.MaxAttendees > 0 && party.AttendeeCount >= party.MaxAttendees {
		isFull = true
	}

	return models.PartyListItem{
		ID:         party.ID,
		Title:      party.Title,
		Type:       party.Type,
		Location:   party.LocationName, // 修正字段名：Location -> LocationName
		Distance:   distance,
		Price:      priceStr,
		Lat:        party.Latitude,
		Lng:        party.Longitude,
		Tags:       tags,
		User:       userName,
		UserAvatar: userAvatar,
		Time:       timeStr,
		Attendees:  party.AttendeeCount,
		IsFull:     isFull,
		Rank:       party.RankLabel,
		Image:      party.CoverImage,
	}
}

// 更新热度分数
func (pc *Party) updateHotScore(partyID int64) {
	var party models.Party
	if err := pc.DB.Where("id = ?", partyID).First(&party).Error; err != nil {
		return
	}

	hotScore := party.AttendeeCount*10 +
		party.LikeCount*5 +
		party.ViewCount*1 +
		party.CommentCount*3

	// 时间衰减：越接近开始时间，热度越高
	now := time.Now()
	if party.StartTime.After(now) {
		daysUntil := party.StartTime.Sub(now).Hours() / 24
		if daysUntil <= 7 {
			// 7天内的派对，热度加成
			hotScore += int64((7 - daysUntil) * 100)
		}
	}

	pc.DB.Model(&models.Party{}).Where("id = ?", partyID).
		UpdateColumn("hot_score", hotScore)
}
