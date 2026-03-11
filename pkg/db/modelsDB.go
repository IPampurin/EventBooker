package db

import "time"

type Event struct {
	ID                int       // id события
	Name              string    // название события
	DateEvent         time.Time // время проведения события
	BookingTTLMinutes int       // время в минутах, на которое можно сделать бронь
	TotalSeats        int       // количество мест всего
	FreeSeats         int       // количество свободных мест
	BookingPrice      int       // стоимость, которую надо внести при бронировании
}

type User struct {
	ID    int    // id пользователя
	Name  string // имя пользователя
	Email string // почта для внешней идентификации
}

type Booking struct {
	ID          int        // id брони
	EventID     int        // id мероприятия
	UsersID     int        // id пользователя, сделавшего бронь
	Status      string     // статус: pending, confirmed, cancelled
	CreatedAt   time.Time  // время создания брони
	ExpiresAt   time.Time  // время, до которого нужно подтвердить (для pending)
	ConfirmedAt *time.Time // время подтверждения оплатой (nil, если ещё не подтверждено)
}
