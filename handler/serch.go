package handler

import (
	"Hyper/config"
	"Hyper/middleware"
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
	authorize := middleware.Auth([]byte(s.Config.Jwt.Secret))

	serchGroup := r.Group("/v1/search")
	serchGroup.GET("/", authorize, context.Wrap(s.Globalserch))
	serchGroup.GET("/history", authorize, context.Wrap(s.GetSearchHistory))
	serchGroup.DELETE("/history", authorize, context.Wrap(s.DeleteSearchKeyword))

}

func (s *SearchHandler) DeleteSearchKeyword(c *gin.Context) error {
	var req types.GlobalSearchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}
	err := s.Serch.DeleteSearchKeyword(c, c.GetInt("user_id"), req.Keyword)
	if err != nil {
		return err

	}
	response.Success(c, "ok")
	return nil
}

func (s *SearchHandler) GetSearchHistory(c *gin.Context) error {
	res, err := s.Serch.GetSearchHistory(c, c.GetInt("user_id"))
	if err != nil {
		return err

	}
	response.Success(c, res)
	return nil
}

func (s *SearchHandler) Globalserch(c *gin.Context) error {
	var req types.GlobalSearchReq
	if err := c.ShouldBindQuery(&req); err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}

	if req.Keyword == "" {
		return response.NewError(http.StatusInternalServerError, "搜索关键词不能为空")
	}

	res, err := s.Serch.GlobalSerch(c.Request.Context(), req)
	s.Serch.SaveSearchHistory(c, c.GetInt("user_id"), req.Keyword)
	if err != nil {
		return response.NewError(http.StatusInternalServerError, err.Error())
	}
	response.Success(c, res)
	return nil
}
