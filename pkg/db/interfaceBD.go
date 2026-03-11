package db

import (
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

	// GetEvents - получение всех предстоящих мероприятий с информацией о свободных местах
	GetEvents() ([]*Event, error)

	// GetEventByID - получение события по id
	GetEventByID(id int) (*Event, error)
}

// BookerTableMethods - управление таблицей бронирования
type BookerTableMethods interface {

	// SeatReserver - бронирование места на мероприятии
	SeatReserver(eventID, userID int, createdAt time.Time) (int, error)

	// GetEventReserveOfUser - получение данных о брони пользователя на мероприятии (да, один юзер - одно место)
	GetEventReserveOfUser(eventID, userID int) (int, error)

	// ReserveConfirmer - метод оплаты/подтверждения бронирования
	ReserveConfirmer(bookingID int) error

	// CancelBooking - отмена брони
	CancelBooking(bookingID int) error
}

// ManageUsersTable - управление таблицей с пользователями
type ManageUsersTable interface {

	// RegisterUser - метод для регистрации пользователя
	RegisterUser(name, email string) (int, error)
}
