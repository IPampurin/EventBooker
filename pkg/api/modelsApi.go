package api

import "time"

// createEventRequest тело запроса для создания мероприятия
type createEventRequest struct {
	Name              string    `json:"name" binding:"required"`
	Date              time.Time `json:"date" binding:"required"`
	BookingTTLMinutes int       `json:"bookingTTLMinutes" binding:"required,min=1"`
	TotalSeats        int       `json:"totalSeats" binding:"required,min=1"`
	BookingPrice      int       `json:"bookingPrice" binding:"min=0"`
}

// confirmRequest тело запроса для подтверждения брони
type confirmRequest struct {
	BookingID int `json:"bookingId" binding:"required"`
}

// bookRequest тело запроса для бронирования
type bookRequest struct {
	UserID int `json:"userId" binding:"required"`
}

// registerRequest тело запроса для регистрации пользователя
type registerRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}
