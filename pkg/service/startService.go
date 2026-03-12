package service

import "github.com/IPampurin/EventBooker/pkg/db"

type Service struct {
	event   db.EventsTableMethods
	booking db.BookingTableMethods
	user    db.UsersTableMethods
}
