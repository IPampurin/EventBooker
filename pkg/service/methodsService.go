package service

import (
	"context"
	"fmt"
	"time"

	"github.com/IPampurin/EventBooker/pkg/domain"
)

// EventCreater - создание мероприятия
func (s *Service) EventCreater(ctx context.Context, name string, date time.Time, bookingTTLMinutes, totalSeats, bookingPrice int) (int, error) {

	id, err := s.storage.EventCreater(ctx, name, date, bookingTTLMinutes, totalSeats, bookingPrice)
	if err != nil {
		return 0, fmt.Errorf("ошибка EventCreater при создании события: %w", err)
	}

	return id, nil
}

// GetEvents - получение всех предстоящих мероприятий с информацией о свободных местах
func (s *Service) GetEvents(ctx context.Context) ([]*domain.Event, error) {

	events, err := s.storage.GetEvents(ctx)
	if err != nil {
		return nil, fmt.Errorf("ошибка GetEvents при получении событий: %w", err)
	}

	return events, nil
}

// GetEventByID - получение события по id
func (s *Service) GetEventByID(ctx context.Context, id int) (*domain.Event, error) {

	event, err := s.storage.GetEventByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("ошибка GetEventByID при получении события: %w", err)
	}

	return event, nil
}

// SeatReserver - бронирование места на мероприятии
func (s *Service) SeatReserver(ctx context.Context, eventID, userID int, createdAt time.Time) (int, error) {

}

// GetEventReserveOfUser - получение данных о брони пользователя на мероприятии (да, один юзер - одно место)
func (s *Service) GetEventReserveOfUser(ctx context.Context, eventID, userID int) (int, error) {

}

// ReserveConfirmer - метод оплаты/подтверждения бронирования
func (s *Service) ReserveConfirmer(ctx context.Context, bookingID int) error {

}

// CancelBooking - отмена брони
func (s *Service) CancelBooking(ctx context.Context, bookingID int) error {

}

// RegisterUser - метод для регистрации пользователя
func (s *Service) RegisterUser(ctx context.Context, name, email string) (int, error) {

}
