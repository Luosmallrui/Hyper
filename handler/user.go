package handler

import (
	"Hyper/config"
	"Hyper/middleware"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	"fmt"
	"io"
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
	userID, _ := context.GetUserID(c)

	header, err := c.FormFile("image")
	if err != nil {
		return response.NewError(400, "请选择图片")
	}

	//  大小校验（10MB）
	if header.Size > 10<<20 {
		return response.NewError(400, "图片不能超过10MB")
	}

	file, err := header.Open()
	if err != nil {
		return response.NewError(400, "读取文件失败")
	}
	defer file.Close()

	buf := make([]byte, 512)
	_, _ = file.Read(buf)
	file.Seek(0, io.SeekStart)

	contentType := http.DetectContentType(buf)
	switch contentType {
	case "image/jpeg", "image/png", "image/webp":
	default:
		return response.NewError(400, "不支持的图片格式")
	}

	objectKey := fmt.Sprintf(
		"avatars/%02d/%d/%s%s",
		userID%100,
		userID,
		uuid.NewString()[:8],
		path.Ext(header.Filename),
	)

	if err := u.OssService.UploadReader(c.Request.Context(), file, objectKey); err != nil {
		return response.NewError(500, "上传云端失败")
	}
	fullUrl := fmt.Sprintf("https://cdn.hypercn.cn/%s", objectKey)
	response.Success(c, types.UploadAvatarRes{Url: fullUrl})
	return nil
}
