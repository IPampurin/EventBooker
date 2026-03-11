package db

import (
	"context"
	"fmt"
	"time"

	"github.com/IPampurin/EventBooker/pkg/domain"
	"github.com/jackc/pgx/v5"
)

// EventCreater - создание мероприятия
func (d *DataBase) EventCreater(ctx context.Context, name string, date time.Time, bookingTTLMinutes, totalSeats, bookingPrice int) (int, error) {

	query := `   INSERT INTO events(name, date_event, booking_ttl_minutes, total_seats, free_seats, booking_price)
	             VALUES ($1, $2, $3, $4, $5, $6)
			  RETURNING id`

	var id int
	err := d.Pool.QueryRow(ctx, query, name, date, bookingTTLMinutes, totalSeats, totalSeats, bookingPrice).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("ошибка EventCreater добавления записи в events: %w", err)
	}

	return id, nil
}

// GetEvents - получение всех предстоящих мероприятий с информацией о свободных местах
func (d *DataBase) GetEvents(ctx context.Context) ([]*Event, error) {

	query := `SELECT id, name, date_event, booking_ttl_minutes, total_seats, free_seats, booking_price
	            FROM events
			   ORDER BY date_event`

	rows, err := d.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка GetEvents при получении записей из events: %w", err)
	}
	defer rows.Close()

	events := make([]*Event, 0)
	for rows.Next() {
		e := &Event{}
		err := rows.Scan(
			&e.ID,
			&e.Name,
			&e.DateEvent,
			&e.BookingTTLMinutes,
			&e.TotalSeats,
			&e.FreeSeats,
			&e.BookingPrice,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка GetEvents при сканировании записи из events: %w", err)
		}

		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка GetEvents при итерации по записям из events: %w", err)
	}

	// TODO: переделать на возврат доменной модели

	return events, nil
}

// GetEventByID - получение события по id
func (d *DataBase) GetEventByID(ctx context.Context, id int) (*Event, error) {

	query := `SELECT id, name, date_event, booking_ttl_minutes, total_seats, free_seats, booking_price
	            FROM events
			   WHERE id = $1`

	e := &Event{}
	err := d.Pool.QueryRow(ctx, query, id).Scan(
		&e.ID,
		&e.Name,
		&e.DateEvent,
		&e.BookingTTLMinutes,
		&e.TotalSeats,
		&e.FreeSeats,
		&e.BookingPrice,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка GetEventByID получения записи в events", err)
	}

	// TODO: переделать на возврат доменной модели

	return e, nil
}

// SeatReserver - бронирование места на мероприятии
func (d *DataBase) SeatReserver(ctx context.Context, eventID, userID int, createdAt, expiresAt time.Time) (int, error) {

	tx, err := d.Pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("ошибка SeatReserver при начале транзакции: %w", err)
	}
	defer tx.Rollback(ctx)

	// проверяем наличие свободных мест и блокируем строку
	var freeSeats int
	checkQueery := `SELECT free_seats
	                  FROM events
					 WHERE id = $1 FOR UPDATE`

	err = tx.QueryRow(ctx, checkQueery, eventID).Scan(&freeSeats)
	if err != nil {
		return 0, fmt.Errorf("ошибка SeatReserver при проверке free_seats в транзакции: %w", err)
	}
	if freeSeats <= 0 {
		return 0, fmt.Errorf("ошибка SeatReserver: нет доступных свободных мест")
	}

	// уменьшаем количество свободных мест
	updateQuery := `UPDATE events
	                   SET free_seats = free_seats - 1
					 WHERE id = $1`

	_, err = tx.Exec(ctx, updateQuery, eventID)
	if err != nil {
		return 0, fmt.Errorf("ошибка SeatReserver при обновлении количества свободных мест: %w", err)
	}

	// создаём бронь
	query := `   INSERT INTO bookings(event_id, user_id, status, created_at, expires_at)
	             VALUES ($1, $2, $3, $4)
			  RETURNING id`

	var id int
	err = tx.QueryRow(ctx, query, eventID, userID, createdAt, expiresAt).Scan(&id)
	if err != nil {
		return id, fmt.Errorf("ошибка SeatReserver добавления записи о новой брони: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("ошибка SeatReserver коммита транзакции: %w", err)
	}

	return id, nil
}

// GetEventReserveOfUser - получение данных о брони пользователя на мероприятии (0 если нет) (да, один юзер - одно место)
func (d *DataBase) GetEventReserveOfUser(ctx context.Context, eventID, userID int) (int, error) {

	query := `SELECT id
	          FROM bookings
			 WHERE event_id = $1 AND user_id = $2`

	var id int
	err := d.Pool.QueryRow(ctx, query, eventID, userID).Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("ошибка GetEventReserveOfUser при получении id брони: %w", err)
	}

	return id, nil
}

// ReserveConfirmer - подтверждение бронирования (оплата)
func (d *DataBase) ReserveConfirmer(ctx context.Context, bookingID int) error {

	query := `UPDATE bookings
	             SET status = $1, confirmed_at = NOW()
	           WHERE id = $2 AND status = $3`

	cmd, err := d.Pool.Exec(ctx, query, domain.BookingStatusConfirmed, bookingID, domain.BookingStatusPending)
	if err != nil {
		return fmt.Errorf("ошибка ReserveConfirmer при обновлении статуса брони: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("ошибка ReserveConfirmer: бронь не найдена или уже подтверждена")
	}

	return nil
}

// CancelBooking - отмена брони (освобождение места)
func (d *DataBase) CancelBooking(ctx context.Context, bookingID int) error {

	tx, err := d.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("ошибка CancelBooking при начале транзакции: %w", err)
	}
	defer tx.Rollback(ctx)

	// получаем event_id и текущий статус (блокируем строку брони)
	var eventID int
	var status string

	selectQuery := `SELECT event_id, status
	                  FROM bookings
					 WHERE id = $1 FOR UPDATE`

	err = tx.QueryRow(ctx, selectQuery, bookingID).Scan(&eventID, &status)
	if err != nil {
		return fmt.Errorf("ошибка CancelBooking при поиске брони: %w", err)
	}
	if status != domain.BookingStatusPending && status != domain.BookingStatusConfirmed {
		// уже отменена — ничего не делаем
		return nil
	}

	// обновляем статус брони
	updateBookingQuery := `UPDATE bookings
	                          SET status = $1
						    WHERE id = $2`

	_, err = tx.Exec(ctx, updateBookingQuery, domain.BookingStatusCancelled, bookingID)
	if err != nil {
		return fmt.Errorf("ошибка CancelBooking при обновлении статуса брони в транзакции: %w", err)
	}

	// увеличиваем количество свободных мест
	updateEventQuery := `UPDATE events
	                        SET free_seats = free_seats + 1
						  WHERE id = $1`

	_, err = tx.Exec(ctx, updateEventQuery, eventID)
	if err != nil {
		return fmt.Errorf("ошибка CancelBooking при обновлении количества доступных мест в events в транзакции: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("ошибка CancelBooking коммита транзакции: %w", err)
	}

	return nil
}

// RegisterUser - метод для регистрации пользователя
func (d *DataBase) RegisterUser(ctx context.Context, name, email string) (int, error) {

	query := `   INSERT INTO users (name, email)
	             VALUES ($1, $2)
			  RETURNING id`

	var id int
	err := d.Pool.QueryRow(ctx, query, name, email).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("ошибка RegisterUser при добавлении пользователя: %w", err)
	}

	return id, nil
}
