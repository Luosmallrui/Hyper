package models

type SearchUser struct {
	ID       uint64 `gorm:"column:id"`
	Nickname string `gorm:"column:nickname"`
	Avatar   string `gorm:"column:avatar"`
}

func (SearchUser) TableName() string {
	return "users"
}

type SearchPost struct {
	ID      uint64 `gorm:"column:id"`
	Title   string `gorm:"column:title"`
	Content string `gorm:"column:content"`
}

func (SearchPost) TableName() string {
	return "posts"
}
