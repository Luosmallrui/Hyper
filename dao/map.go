package dao

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type MapDao struct{}

func NewMapDao() *MapDao {
	return &MapDao{}
}

// GetMapData 从 config/map.json 读取地图数据
func (m *MapDao) GetMapData() (interface{}, error) {
	// 获取 map.json 文件路径
	configPath := filepath.Join("config", "map.json")

	// 读取文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// 解析 JSON
	var mapData interface{}
	if err := json.Unmarshal(data, &mapData); err != nil {
		return nil, err
	}

	return mapData, nil
}
