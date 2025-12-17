package response

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type BizError struct {
	Code int
	Msg  string
}

func (e *BizError) Error() string {
	return e.Msg
}

func NewError(code int, msg string) *BizError {
	return &BizError{
		Code: code,
		Msg:  msg,
	}
}

func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				c.JSON(http.StatusInternalServerError, Response{
					Code: 500,
					Msg:  "系统异常",
				})
				c.Abort()
			}
		}()

		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			if be, ok := err.(*BizError); ok {
				Fail(c, be.Code, be.Msg)
			} else {
				Fail(c, 500, err.Error())
			}
			c.Abort()
		}
	}
}

func Abort(c *gin.Context, httpStatus int, msg string) {
	c.AbortWithStatusJSON(httpStatus, Response{
		Code: httpStatus,
		Msg:  msg,
		Data: nil,
	})
}
