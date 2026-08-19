package response

import (
	"github.com/gin-gonic/gin"
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

func Abort(c *gin.Context, httpStatus int, msg string) {
	c.AbortWithStatusJSON(httpStatus, Response{
		Code: httpStatus,
		Msg:  msg,
		Data: nil,
	})
}
