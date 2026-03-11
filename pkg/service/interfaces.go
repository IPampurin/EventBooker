package service

import (
	"context"
	"time"

	"github.com/IPampurin/EventBooker/pkg/domain"
)

// AdminMethods - админское управление событиями
type AdminMethods interface {

	// EventCreater - создание мероприятия
	EventCreater(ctx context.Context, name string, date time.Time, bookingTTLMinutes, totalSeats, bookingPrice int) (int, error)
}

// EventMethods - пользовательская работа с событиями
type EventMethods interface {

	// GetEvents - получение всех предстоящих мероприятий с информацией о свободных местах
	GetEvents(ctx context.Context) ([]*domain.Event, error)

	// GetEventByID - получение события по id
	GetEventByID(ctx context.Context, id int) (*domain.Event, error)
}

// BookerMethods - управление бронированием
type BookerMethods interface {

	// SeatReserver - бронирование места на мероприятии
	SeatReserver(ctx context.Context, eventID, userID int, createdAt time.Time) (int, error)

	// GetEventReserveOfUser - получение данных о брони пользователя на мероприятии (да, один юзер - одно место)
	GetEventReserveOfUser(ctx context.Context, eventID, userID int) (int, error)

	// ReserveConfirmer - метод оплаты/подтверждения бронирования
	ReserveConfirmer(ctx context.Context, bookingID int) error

	// CancelBooking - отмена брони
	CancelBooking(ctx context.Context, bookingID int) error
}

// ManageUsers - управление пользователями
type ManageUsers interface {

	// RegisterUser - метод для регистрации пользователя
	RegisterUser(ctx context.Context, name, email string) (int, error)
}
