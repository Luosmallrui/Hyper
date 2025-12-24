package handler

import (
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"

	"github.com/gin-gonic/gin"
)

type Map struct {
	MapService service.IMapService
	OssService service.IOssService
}

func (m *Map) RegisterRouter(r gin.IRouter) {
	mapGroup := r.Group("/")
	mapGroup.GET("/map", context.Wrap(m.GetMap))
	mapGroup.GET("/test", context.Wrap(m.Test))
}
func (m *Map) Test(c *gin.Context) error {
	mapData, err := m.OssService.ListBuckets(c.Request.Context())
	if err != nil {
		return err
	}
	response.Success(c, mapData)
	return nil
}

func (m *Map) GetMap(c *gin.Context) error {
	mapData, err := m.MapService.GetMapData()
	if err != nil {
		response.Fail(c, 500, "获取地图数据失败")
		return err
	}
	response.Success(c, mapData)
	return nil
}
