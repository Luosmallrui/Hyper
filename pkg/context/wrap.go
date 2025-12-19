package context

import (
	"Hyper/pkg/response"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	CtxUserID = "user_id"
	CtxOpenID = "openid"
)

type HandlerFunc func(*gin.Context) error

func Wrap(h func(*gin.Context) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := h(c); err != nil {

			// 如果已经写过响应，直接返回
			if c.Writer.Written() {
				return
			}
			// 业务错误
			var be *response.BizError
			if errors.As(err, &be) {
				c.JSON(http.StatusOK, response.Response{
					Code: be.Code,
					Msg:  be.Msg,
				})
				return
			}
			c.JSON(http.StatusInternalServerError, response.Response{
				Code: 500,
				Msg:  err.Error(),
			})
		}
	}
}

func GetUserID(c *gin.Context) (int64, error) {
	v, ok := c.Get(CtxUserID)
	if !ok {
		return 0, errors.New("user_id 不存在")
	}
	switch uid := v.(type) {
	case int64:
		return uid, nil
	case uint:
		return int64(uid), nil
	case int:
		return int64(uid), nil
	case float64:
		return int64(uid), nil
	default:
		return 0, errors.New("user_id 类型不支持")
	}
}
