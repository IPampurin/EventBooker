package service

import (
	"time"

	"github.com/IPampurin/EventBooker/pkg/domain"
)

// AdminMethods - админское управление событиями
type AdminMethods interface {

	// EventCreater - создание мероприятия
	EventCreater(name string, date time.Time, bookingTTLMinutes, totalSeats, bookingPrice int) (int, error)
}

// EventMethods - пользовательская работа с событиями
type EventMethods interface {

	// GetEvents - получение всех предстоящих мероприятий с информацией о свободных местах
	GetEvents() ([]*domain.Event, error)

	// GetEventByID - получение события по id
	GetEventByID(id int) (*domain.Event, error)
}

// BookerMethods - управление бронированием
type BookerMethods interface {

	// SeatReserver - бронирование места на мероприятии
	SeatReserver(eventID, userID int, createdAt time.Time) (int, error)

	// GetEventReserveOfUser - получение данных о брони пользователя на мероприятии (да, один юзер - одно место)
	GetEventReserveOfUser(eventID, userID int) (int, error)

	// ReserveConfirmer - метод оплаты/подтверждения бронирования
	ReserveConfirmer(bookingID int) error

	// CancelBooking - отмена брони
	CancelBooking(bookingID int) error
}

// ManageUsers - управление пользователями
type ManageUsers interface {

	// RegisterUser - метод для регистрации пользователя
	RegisterUser(name, email string) (int, error)
}
