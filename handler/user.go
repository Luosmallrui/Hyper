package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	"fmt"
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type User struct {
	Config      *config.Config
	UserService service.IUserService
	OssService  service.IOssService
}

func (u *User) RegisterRouter(r gin.IRouter) {
	authorize := middleware.Auth([]byte(u.Config.Jwt.Secret))
	g := r.Group("/v1/user")
	g.Use(authorize)
	g.POST("/info", context.Wrap(u.UpdateUserInfo))
	g.POST("/avatar", context.Wrap(u.UploadAvatar))

}
func (u *User) UpdateUserInfo(c *gin.Context) error {
	userID, err := context.GetUserID(c) // 这里的 userID 是 int
	if err != nil {
		return response.NewError(http.StatusUnauthorized, "未登录")
	}

	var req types.UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, "参数格式错误")
	}

	if err = u.UserService.Update(c.Request.Context(), int(userID), &req); err != nil {
		return response.NewError(http.StatusInternalServerError, "更新失败: "+err.Error())
	}
	response.Success(c, "更新成功")
	return nil
}
func (u *User) UploadAvatar(c *gin.Context) error {
	userID, _ := context.GetUserID(c) // 获取当前用户ID

	header, err := c.FormFile("image")
	if err != nil {
		return response.NewError(400, "请选择图片")
	}

	// 1. 散列目录设计：按 userID % 100 分层
	// 2. 唯一文件名：UUID + 时间戳
	// 路径格式：avatars/hash/userID/uuid.ext
	objectKey := fmt.Sprintf("avatars/%02d/%d/%s%s",
		userID%100,
		userID,
		uuid.NewString()[:8],
		path.Ext(header.Filename),
	)

	file, _ := header.Open()
	defer file.Close()

	// 执行上传
	if err := u.OssService.UploadReader(c.Request.Context(), file, objectKey); err != nil {
		return response.NewError(500, "上传云端失败")
	}

	// 返回图片地址，前端拿到这个 Url 后，再调用 UpdateUserInfo 接口
	fullUrl := fmt.Sprintf("https://%s.%s/%s", u.Config.Oss.Bucket, u.Config.Oss.Endpoint, objectKey)
	response.Success(c, gin.H{
		"url": fullUrl,
	})
	return nil
}
