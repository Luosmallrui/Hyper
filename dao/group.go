package dao

import "gorm.io/gorm"

type GroupDAO struct {
	db *gorm.DB
}

func NewGroupDAO(db *gorm.DB) *GroupDAO {
	return &GroupDAO{db: db}
}

// 查询群成员
func (d *GroupDAO) GetGroupMembers(groupID string) ([]string, error) {
	var users []string
	err := d.db.
		Table("im_group_member").
		Select("user_id").
		Where("group_id = ?", groupID).
		Scan(&users).Error
	return users, err
}
