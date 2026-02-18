package service

import (
	"Hyper/dao"

	"gorm.io/gorm"
)

type MapService struct {
	MapDao *dao.MapDao
	DB     *gorm.DB
}

var _ IMapService = (*MapService)(nil)

type IMapService interface {
	GetMapData() (interface{}, error)
}

// GetMapData 获取地图数据
func (s *MapService) GetMapData() (interface{}, error) {
	return s.MapDao.GetMapData()
}

//// 城市默认为1成都，后续推进可以增加城市选择功能
//func (s *MapService) GetDistrictsList(ctx context.Context, cityID int) ([]models.District, error) {
//	var districts []models.District
//	err := s.DB.Model(&models.District{}).
//		Where("city_id = ?", cityID).
//		Order("sorted_id asc").
//		Find(&districts).Error
//	if err != nil {
//		return nil, errors.New("获取行政区列表失败" + err.Error())
//	}
//	return districts, err
//}
//
//func (s *MapService) GetDistrictsListDetail(ctx context.Context, req types.GetDistrictDetailRequest) (types.GetDistrictDetailResponse, error) {
//	var areas []models.Areas
//	db := s.DB.Model(&models.Areas{}).
//		Where("district_id = ? AND is_activate = ?", req.DistrictID, 1)
//
//	if req.Cursor > 0 {
//		db = db.Where("id > ?", req.Cursor)
//	}
//
//	err := db.Order("id ASC").Limit(req.Limit).Find(&areas).Error
//	if err != nil {
//		return types.GetDistrictDetailResponse{}, errors.New("获取行政区详情失败" + err.Error())
//	}
//
//	resp := types.GetDistrictDetailResponse{
//		List:    areas,
//		HasMore: len(areas) == req.Limit,
//	}
//
//	if len(resp.List) > 0 {
//		resp.NextCursor = resp.List[len(resp.List)-1].ID
//	}
//	return resp, nil
//}
