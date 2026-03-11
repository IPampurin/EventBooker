package db

import (
	"context"
	"time"
)

// StorageDB - интерфейс локальной БД
type StorageDB interface {
	EventsTableMethods
	BookerTableMethods
	ManageUsersTable
}

// EventsTableMethods - методы для работы с таблицей events
type EventsTableMethods interface {

	// EventCreater - создание мероприятия
	EventCreater(ctx context.Context, name string, date time.Time, bookingTTLMinutes, totalSeats, bookingPrice int) (int, error)

	// GetEvents - получение всех предстоящих мероприятий с информацией о свободных местах
	GetEvents(ctx context.Context) ([]*Event, error)

	// GetEventByID - получение события по id
	GetEventByID(ctx context.Context, id int) (*Event, error)
}

// BookerTableMethods - управление таблицей бронирования
type BookerTableMethods interface {

	// SeatReserver - бронирование места на мероприятии
	SeatReserver(ctx context.Context, eventID, userID int, createdAt, expiresAt time.Time) (int, error)

	// GetEventReserveOfUser - получение данных о брони пользователя на мероприятии (да, один юзер - одно место)
	GetEventReserveOfUser(ctx context.Context, eventID, userID int) (int, error)

	// ReserveConfirmer - метод оплаты/подтверждения бронирования
	ReserveConfirmer(ctx context.Context, bookingID int) error

	// CancelBooking - отмена брони
	CancelBooking(ctx context.Context, bookingID int) error
}

// ManageUsersTable - управление таблицей с пользователями
type ManageUsersTable interface {

	// RegisterUser - метод для регистрации пользователя
	RegisterUser(ctx context.Context, name, email string) (int, error)
}
