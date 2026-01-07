package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
	"Hyper/pkg/jwt"
	"Hyper/pkg/response"
	"Hyper/pkg/utils"
	"Hyper/service"
	"Hyper/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Auth struct {
	Config         *config.Config
	UserService    service.IUserService
	WeChatService  service.IWeChatService
	OssService     service.IOssService
	FollowService  service.IFollowService
	LikeService    service.ILikeService
	CollectService service.ICollectService
}

func (u *Auth) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(u.Config.Jwt.Secret))
	auth := r.Group("/")
	auth.POST("/auth/wx-login", context.Wrap(u.Login))                  // 微信登录
	auth.POST("/auth/bind-phone", authorize, context.Wrap(u.BindPhone)) //微信获取手机号
	auth.GET("/token", context.Wrap(u.GetToken))
	// auth.GET("/test1", context.Wrap(u.test1))
}

// func (u *Auth) test1(c *gin.Context) error {
// 	userid := 1
// 	following, err := u.FollowService.GetFollowingCount(c.Request.Context(), uint64(userid))
// 	if err != nil {
// 		following = 0
// 	}

// 	// 获取粉丝数
// 	follower, err := u.FollowService.GetFollowerCount(c.Request.Context(), uint64(userid))
// 	if err != nil {
// 		follower = 0
// 	}

// 	// 获取用户帖子的总点赞数 + 总收藏数
// 	totalLikes, err := u.LikeService.GetUserTotalLikes(c.Request.Context(), uint64(userid))
// 	if err != nil {
// 		totalLikes = 0
// 	}

// 	totalCollects, err := u.CollectService.GetUserTotalCollects(c.Request.Context(), uint64(userid))
// 	if err != nil {
// 		totalCollects = 0
// 	}

// 	rep := types.UserProfileResp{
// 		Stats: types.UserStats{
// 			Following: following,
// 			Follower:  follower,
// 			Likes:     totalLikes + totalCollects,
// 		},
// 	}
// 	response.Success(c, rep)
// 	return nil
// }

func (u *Auth) GetToken(c *gin.Context) error {
	token, err := jwt.GenerateToken([]byte(u.Config.Jwt.Secret), 1, "XXX")
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, token)
	return nil
}

func (u *Auth) Login(c *gin.Context) error {
	var req types.WxLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "参数格式错误")
	}

	if req.LoginCode == "" {
		return response.NewError(http.StatusInternalServerError, "login_code 不能为空")
	}

	wxResp, err := u.WeChatService.Code2Session(c.Request.Context(), req.LoginCode)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}

	user, err := u.UserService.GetOrCreateByOpenID(c.Request.Context(), wxResp.OpenID)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}

	token, err := jwt.GenerateToken([]byte(u.Config.Jwt.Secret), uint(user.Id), user.OpenID)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}

	// 获取关注数
	following, err := u.FollowService.GetFollowingCount(c.Request.Context(), uint64(user.Id))
	if err != nil {
		following = 0
	}

	// 获取粉丝数
	follower, err := u.FollowService.GetFollowerCount(c.Request.Context(), uint64(user.Id))
	if err != nil {
		follower = 0
	}

	// 获取用户帖子的总点赞数 + 总收藏数
	totalLikes, err := u.LikeService.GetUserTotalLikes(c.Request.Context(), uint64(user.Id))
	if err != nil {
		totalLikes = 0
	}

	totalCollects, err := u.CollectService.GetUserTotalCollects(c.Request.Context(), uint64(user.Id))
	if err != nil {
		totalCollects = 0
	}

	rep := types.UserProfileResp{
		User: types.UserBasicInfo{
			UserID:      utils.GenHashID(u.Config.Jwt.Secret, user.Id),
			Nickname:    user.Nickname,
			PhoneNumber: user.Mobile,
			AvatarURL:   user.Avatar,
		},
		Stats: types.UserStats{
			Following: following,
			Follower:  follower,
			Likes:     totalLikes + totalCollects,
		},
		Token: token,
	}
	response.Success(c, rep)
	return nil
}

func (u *Auth) BindPhone(c *gin.Context) error {
	userId, err := context.GetUserID(c)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	var req types.BindPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.PhoneCode == "" {
		return response.NewError(http.StatusInternalServerError, "phone_code 不能为空")
	}
	userPhoneNumber, err := u.WeChatService.GetUserPhoneNumber(req.PhoneCode)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	phone := userPhoneNumber
	err = u.UserService.UpdateMobile(c.Request.Context(), int(userId), phone)
	if err != nil {
		return response.NewError(500, err.Error())
	}

	response.Success(c, types.BindPhoneRep{PhoneNumber: phone})
	return nil
}
