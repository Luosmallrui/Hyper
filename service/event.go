package service

import (
	"Hyper/dao"
	"Hyper/models"
	"Hyper/types"
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type EventService struct {
	EventDAO  *dao.EventDAO
	TicketDAO *dao.EventTicketDAO
	DB        *gorm.DB
}

var _ IEventService = (*EventService)(nil)

type IEventService interface {
	CreateEvent(ctx context.Context, userID int, req types.CreateEventRequest) (int64, error)
	GetEventDetail(ctx context.Context, eventID int64) (*types.EventDetail, error)
	GetEventList(ctx context.Context, page, pageSize int) ([]models.EventListItem, int64, error)
	GetMyEvents(ctx context.Context, userID int, page, pageSize int) ([]models.EventListItem, int64, error)
}

func (s *EventService) CreateEvent(ctx context.Context, userID int, req types.CreateEventRequest) (int64, error) {
	startAt, err := time.Parse("2006-01-02T15:04", req.StartAt)
	if err != nil {
		return 0, errors.New("开始时间格式无效")
	}
	endAt, err := time.Parse("2006-01-02T15:04", req.EndAt)
	if err != nil {
		return 0, errors.New("结束时间格式无效")
	}
	if endAt.Before(startAt) {
		return 0, errors.New("结束时间不能早于开始时间")
	}

	var eventID int64
	err = s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		event := &models.Event{
			UserID:            userID,
			Title:             req.Title,
			ShareTitle:        req.ShareTitle,
			StartAt:           startAt,
			EndAt:             endAt,
			Summary:           req.Summary,
			SummaryHTML:       req.SummaryHTML,
			StrongRealName:    req.StrongRealName,
			MinorProtection:   req.MinorProtection,
			AddressMode:       req.Address.Mode,
			Province:          req.Address.Province,
			City:              req.Address.City,
			District:          req.Address.District,
			Location:          req.Address.Location,
			DetailPoster:      req.Posters.DetailPoster,
			DetailLongPoster:  req.Posters.DetailLongPoster,
			SharePoster:       req.Posters.SharePoster,
			GroupPoster:       req.Posters.GroupPoster,
			PromoterConfirmed: req.Qualification.PromoterConfirmed,
			SafetyAgreement:   req.Qualification.SafetyAgreement,
			TicketStatement:   req.Qualification.TicketStatement,
			Notes:             req.Qualification.Notes,
			Status:            "已上架",
		}

		if err := tx.Create(event).Error; err != nil {
			return err
		}

		// 批量创建票券
		tickets := make([]*models.EventTicket, 0, len(req.Tickets))
		for _, t := range req.Tickets {
			ticketStartAt, err := time.Parse("2006-01-02T15:04", t.StartAt)
			if err != nil {
				return errors.New("票券开售时间格式无效")
			}
			ticketEndAt, err := time.Parse("2006-01-02T15:04", t.EndAt)
			if err != nil {
				return errors.New("票券停售时间格式无效")
			}
			tickets = append(tickets, &models.EventTicket{
				EventID:       event.ID,
				Name:          t.Name,
				Price:         t.Price * 100, // 元转分
				Stock:         t.Stock,
				PurchaseLimit: t.PurchaseLimit,
				StartAt:       ticketStartAt,
				EndAt:         ticketEndAt,
				Description:   t.Description,
			})
		}
		if err := tx.Create(tickets).Error; err != nil {
			return err
		}

		eventID = event.ID
		return nil
	})

	return eventID, err
}

func (s *EventService) GetEventDetail(ctx context.Context, eventID int64) (*types.EventDetail, error) {
	event, err := s.EventDAO.FindById(ctx, eventID)
	if err != nil {
		return nil, err
	}

	var tickets []*models.EventTicket
	err = s.DB.WithContext(ctx).
		Where("event_id = ?", eventID).
		Find(&tickets).Error
	if err != nil {
		return nil, err
	}

	eventTickets := make([]models.EventTicket, 0, len(tickets))
	for _, t := range tickets {
		eventTickets = append(eventTickets, *t)
	}

	return &types.EventDetail{
		Event:   *event,
		Tickets: eventTickets,
	}, nil
}

func (s *EventService) GetEventList(ctx context.Context, page, pageSize int) ([]models.EventListItem, int64, error) {
	var total int64
	err := s.DB.WithContext(ctx).
		Model(&models.Event{}).
		Where("status = ?", "已上架").
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	var events []models.Event
	err = s.DB.WithContext(ctx).
		Where("status = ?", "已上架").
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&events).Error
	if err != nil {
		return nil, 0, err
	}

	list := make([]models.EventListItem, 0, len(events))
	for _, e := range events {
		list = append(list, models.EventListItem{
			ID:          e.ID,
			Title:       e.Title,
			ShareTitle:  e.ShareTitle,
			SharePoster: e.SharePoster,
			StartAt:     e.StartAt,
			EndAt:       e.EndAt,
			Status:      e.Status,
			City:        e.City,
			Location:    e.Location,
			CreatedAt:   e.CreatedAt,
		})
	}

	return list, total, nil
}

func (s *EventService) GetMyEvents(ctx context.Context, userID int, page, pageSize int) ([]models.EventListItem, int64, error) {
	var total int64
	err := s.DB.WithContext(ctx).
		Model(&models.Event{}).
		Where("user_id = ?", userID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	var events []models.Event
	err = s.DB.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&events).Error
	if err != nil {
		return nil, 0, err
	}

	list := make([]models.EventListItem, 0, len(events))
	for _, e := range events {
		list = append(list, models.EventListItem{
			ID:          e.ID,
			Title:       e.Title,
			ShareTitle:  e.ShareTitle,
			SharePoster: e.SharePoster,
			StartAt:     e.StartAt,
			EndAt:       e.EndAt,
			Status:      e.Status,
			City:        e.City,
			Location:    e.Location,
			CreatedAt:   e.CreatedAt,
		})
	}

	return list, total, nil
}
