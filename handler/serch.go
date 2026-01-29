package handler

import (
	"Hyper/config"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	Config *config.Config
	Serch  service.SearchService
}

func (s *SearchHandler) RegisterRouter(r gin.IRouter) {
	//authorize := middleware.Auth([]byte(s.Config.Jwt.Secret))

	serchGroup := r.Group("/v1/search")
	serchGroup.GET("/searchgobal", context.Wrap(s.GlobalSerch))
}

func (s *SearchHandler) GlobalSerch(c *gin.Context) error {
	var req types.GlobalSearchReq
	if err := c.ShouldBindQuery(&req); err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}

	if req.Keyword == "" {
		return response.NewError(http.StatusInternalServerError, "搜索关键词不能为空")
	}

	res, err := s.Serch.GlobalSerch(c.Request.Context(), req)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, res)
	return nil
}
