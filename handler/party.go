package handler

import (
	"Hyper/models"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Merchant struct {
	DB          *gorm.DB
	UserService service.IUserService
	NoteService service.INoteService
}

func (pc *Merchant) RegisterRouter(r gin.IRouter) {
	m := r.Group("/v1/merchant")
	{
		// 公开接口
		m.GET("/list", context.Wrap(pc.GetPartyList))  // 获取派对列表
		m.GET("/:id", context.Wrap(pc.GetPartyDetail)) // 获取派对详情

		m.POST("/create", pc.CreateMerchant) // 创建商家

		m.POST("/:id/attend", pc.AttendParty)    // 报名
		m.DELETE("/:id/attend", pc.CancelAttend) // 取消报名
	}
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

func (pc *Merchant) GetPartyList(c *gin.Context) error {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	ctx := c.Request.Context()

	query := pc.DB.Model(&models.Merchant{})

	var total int64
	total = 2
	offset := (page - 1) * pageSize
	var merchant []models.Merchant

	if err := query.Offset(offset).Limit(pageSize).Find(&merchant).Error; err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "查询失败", "data": nil})
		return nil
	}

	userIDArr := make([]uint64, 0)
	for _, m := range merchant {
		userIDArr = append(userIDArr, uint64(m.UserID))
	}

	userMap := pc.UserService.BatchGetUserInfo(ctx, userIDArr)

	list := make([]models.MerchantListItem, 0, len(merchant))
	for _, m := range merchant {
		userId := uint64(m.UserID)
		userAvatar := userMap[userId].Avatar
		userName := userMap[userId].Nickname
		list = append(list, models.MerchantListItem{
			ID:           m.ID,
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
func (pc *Merchant) GetPartyDetail(c *gin.Context) error {
	MerchantID := c.Param("id")
	if MerchantID == "" {
		response.Success(c, nil)
	}

	resp := types.MerchantDetail{}
	var marchant models.Merchant
	if err := pc.DB.Where("id = ?", MerchantID).First(&marchant).Error; err != nil {
	}

	resp.AvgPrice = 7600
	resp.Name = marchant.Title
	resp.LocationName = marchant.LocationName
	images := make([]string, 0)
	_ = json.Unmarshal([]byte(marchant.ImagesJSON), &images)
	resp.Images = images

	goods := make([]models.Product, 0)
	if err := pc.DB.Where("party_id = ?", MerchantID).Find(&goods).Error; err != nil {
	}
	resp.Goods = goods

	avatar, nickname, _ := pc.UserService.GetUserAvatar(c.Request.Context(), int64(marchant.UserID))
	resp.UserAvatar = avatar
	resp.UserName = nickname
	resp.IsFollow = true
	resp.BusinessHours = "19:30-次日02:30"
	userId := c.GetInt("user_id")
	var req types.FeedRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "参数错误")
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	notes, _ := pc.NoteService.ListNoteByUser(
		c.Request.Context(),
		req.Cursor,
		req.PageSize,
		userId,
		marchant.UserID,
	)
	resp.Notes = notes

	response.Success(c, resp)
	return nil

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
