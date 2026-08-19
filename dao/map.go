package dao

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type MapDao struct {
	data interface{}
	err  error
}

// NewMapDao 启动时加载并缓存地图数据，避免每次请求都读文件
func NewMapDao() *MapDao {
	m := &MapDao{}
	m.data, m.err = loadMapData()
	return m
}

// GetMapData 返回启动时缓存的地图数据
func (m *MapDao) GetMapData() (interface{}, error) {
	return m.data, m.err
}

// loadMapData 从 config/map.json 读取地图数据
func loadMapData() (interface{}, error) {
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
