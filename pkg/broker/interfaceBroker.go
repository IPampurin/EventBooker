package broker

import (
	"context"

	"github.com/IPampurin/EventBooker/pkg/configuration"
	"github.com/wb-go/wbf/logger"
	"github.com/wb-go/wbf/rabbitmq"
)

type BrokerQ struct {
	*rabbitmq.RabbitClient
}

type BrokerMethods interface {
	InitBroker(ctx context.Context, cfgBroker *configuration.ConfBroker, log logger.Logger) (*BrokerQ, error)
}
