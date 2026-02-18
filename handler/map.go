package handler

import (
	"Hyper/models"
	"Hyper/pkg/context"
	"Hyper/pkg/response"
	"Hyper/service"
	"Hyper/types"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Map struct {
	MapService service.IMapService
	OssService service.IOssService
	Redis      *redis.Client
	DB         *gorm.DB
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
	var districts []models.District
	//city_id 暂时写死为1，后续可以增加城市选择功能
	if err := m.DB.WithContext(c.Request.Context()).
		Where("city_id = ?", 1).
		Order("sort_order asc").
		Find(&districts).Error; err != nil {
		return errors.New("获取行政区列表失败: " + err.Error())
	}

	var dbareas []models.Areas
	if err := m.DB.WithContext(c.Request.Context()).
		Where("is_active = ?", 1).
		Order("sort_order asc").
		Find(&dbareas).Error; err != nil {
		return errors.New("获取街道列表失败: " + err.Error())
	}

	areaMap := make(map[int][]types.Area)

	for _, a := range dbareas {
		areaMap[a.DistrictID] = append(areaMap[a.DistrictID], types.Area{
			ID:         a.ID,
			DistrictID: a.DistrictID,
			Name:       a.Name,
			SortOrder:  a.SortedOrder,
			IsActive:   a.IsActivate == 1,
		})
	}

	tree := make([]types.DistrictTree, 0, len(types.DistrictList))
	for _, d := range districts {
		areas := areaMap[d.ID]
		if areas == nil {
			areas = []types.Area{} // 保证 JSON 输出 [] 而不是 null
		}
		tree = append(tree, types.DistrictTree{
			ID:        d.ID,
			Name:      d.Name,
			SortOrder: d.SortedID,
			Areas:     areas,
		})
	}
	response.Success(c, tree)

	return nil
}
