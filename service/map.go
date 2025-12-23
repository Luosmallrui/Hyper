package service

import (
	"Hyper/dao"
)

type MapService struct {
	MapDao *dao.MapDao
}

var _ IMapService = (*MapService)(nil)

type IMapService interface {
	GetMapData() (interface{}, error)
}

// GetMapData 获取地图数据
func (s *MapService) GetMapData() (interface{}, error) {
	return s.MapDao.GetMapData()
}
