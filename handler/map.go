package handler

import (
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"

	"github.com/gin-gonic/gin"
)

type Map struct {
	MapService *service.MapService
}

func (m *Map) RegisterRouter(r gin.IRouter) {
	mapGroup := r.Group("/")
	mapGroup.GET("/map", context.Wrap(m.GetMap))
}

func (m *Map) GetMap(c *gin.Context) error {
	// 通过 service 层获取地图数据
	mapData, err := m.MapService.GetMapData()
	if err != nil {
		response.Fail(c, 500, "获取地图数据失败")
		return err
	}

	// 返回数据
	response.Success(c, mapData)
	return nil
}
