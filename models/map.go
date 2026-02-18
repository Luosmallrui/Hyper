package models

import (
	"time"
)

type District struct {
	ID         int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name       string    `gorm:"column:name;type:varchar(255);not null" json:"name"`
	CityID     int       `gorm:"column:city_id;type:int;default:1;index" json:"city_id"`
	SortedID   int       `gorm:"column:sorted_id;type:int;default:0" json:"sorted_id"`
	CreateTime time.Time `gorm:"column:create_time;type:datetime;autoCreateTime" json:"create_time"`
}

func (District) TableName() string {
	return "districts"
}

type Areas struct {
	ID          int       `gorm:"primaryKey;column:id" json:"id"`
	DistrictID  int       `gorm:"column:district_id;index" json:"district_id"` // 关联行政区ID
	Name        string    `gorm:"column:name" json:"name"`
	SortedOrder int       `gorm:"column:sorted_order" json:"sorted_order"`
	IsActivate  int       `gorm:"column:is_active;default:1" json:"is_active"` // 1: 激活, 0: 禁用
	CreateAt    time.Time `gorm:"column:create_at" json:"create_at"`
}

func (Areas) TableName() string {
	return "Areas"
}
