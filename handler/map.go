package handler

import (
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Map struct {
	MapService service.IMapService
	OssService service.IOssService
	Redis      *redis.Client
}

func (m *Map) RegisterRouter(r gin.IRouter) {
	mapGroup := r.Group("/v1/districts")
	mapGroup.GET("/map", context.Wrap(m.GetMap))
	mapGroup.GET("/test", context.Wrap(m.Test))
	mapGroup.GET("/tree", context.Wrap(m.GetDistrictTree))
}

func (m *Map) Test(c *gin.Context) error {
	mapData, err := m.OssService.ListBuckets(c.Request.Context())
	if err != nil {
		return err
	}
	fmt.Println(m.Redis, 55)
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

func (m *Map) GetDistrictTree(c *gin.Context) error {
	areaMap := make(map[int][]types.Area, len(types.DistrictList))
	for _, a := range types.AreaList {
		if a.IsActive {
			areaMap[a.DistrictID] = append(areaMap[a.DistrictID], a)
		}
	}

	tree := make([]types.DistrictTree, 0, len(types.DistrictList))
	for _, d := range types.DistrictList {
		areas := areaMap[d.ID]
		if areas == nil {
			areas = []types.Area{} // 保证 JSON 输出 [] 而不是 null
		}
		tree = append(tree, types.DistrictTree{
			ID:        d.ID,
			Name:      d.Name,
			SortOrder: d.SortOrder,
			Areas:     areas,
		})
	}
	response.Success(c, tree)

	return nil
}
