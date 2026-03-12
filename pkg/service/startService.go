package service

import (
	"context"

	"github.com/IPampurin/EventBooker/pkg/broker"
	"github.com/IPampurin/EventBooker/pkg/db"
	"github.com/IPampurin/EventBooker/pkg/zSet"
)

type Service struct {
	storage db.StorageMethods
	zSet    zSet.ZSetMethods
	broker  broker.BrokerMethods
}

func InitService(ctx context.Context, storage db.StorageMethods, clientZSet zSet.ZSetMethods, broker broker.BrokerMethods) *Service {

	// запуск горутины-приёмника канала из брокера

	return &Service{
		storage: storage,
		zSet:    clientZSet,
		broker:  broker,
	}
}

// функция горутина-приёмник из канала из брокера
