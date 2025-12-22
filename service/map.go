package service

import (
	"Hyper/dao"
)

type MapService struct {
	MapDao *dao.MapDao
}

func NewMapService(mapDao *dao.MapDao) *MapService {
	return &MapService{
		MapDao: mapDao,
	}
}

// GetMapData 获取地图数据
func (s *MapService) GetMapData() (interface{}, error) {
	return s.MapDao.GetMapData()
}
