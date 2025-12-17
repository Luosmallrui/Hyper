package dao

import (
	"Hyper/models"
	"gorm.io/gorm"
)

type Admin struct {
	Repo[models.Admin]
}

func NewAdmin(db *gorm.DB) *Admin {
	return &Admin{Repo: NewRepo[models.Admin](db)}
}
