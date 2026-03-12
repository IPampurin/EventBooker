package service

import (
	"context"

	"github.com/IPampurin/EventBooker/pkg/db"
	"github.com/IPampurin/EventBooker/pkg/zSet"
)

type Service struct {
	event   db.EventsTableMethods
	booking db.BookingTableMethods
	user    db.UsersTableMethods
	zSet    zSet.ZSetMethods
}

func InitService(ctx context.Context, storage *db.DataBase, clientZSet *zSet.ClientZSet, overdueCh <-chan int) *Service {

}
