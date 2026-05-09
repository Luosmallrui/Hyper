package dao

import (
	"Hyper/models"

	"gorm.io/gorm"
)

type EventDAO struct {
	Repo[models.Event]
}

func NewEventDAO(db *gorm.DB) *EventDAO {
	return &EventDAO{Repo: NewRepo[models.Event](db)}
}

type EventTicketDAO struct {
	Repo[models.EventTicket]
}

func NewEventTicketDAO(db *gorm.DB) *EventTicketDAO {
	return &EventTicketDAO{Repo: NewRepo[models.EventTicket](db)}
}
