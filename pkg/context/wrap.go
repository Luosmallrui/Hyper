package context

import (
	"Hyper/pkg/response"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
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
