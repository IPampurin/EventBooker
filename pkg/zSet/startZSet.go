package zSet

import (
	"context"
	"fmt"

	"github.com/IPampurin/EventBooker/pkg/configuration"

	"github.com/wb-go/wbf/logger"
	"github.com/wb-go/wbf/redis"
)

// ClientZSet хранит подключение к БД Redis и ключ ZSET для просрочек
type ClientZSet struct {
	*redis.Client
	key string
}

// InitRedis запускает работу с Redis и горутину для отслеживания просроченных бронирований
func InitRedis(ctx context.Context, cfg *configuration.ConfZSet, log logger.Logger) (*ClientZSet, <-chan int, error) {

	// определяем конфигурацию подключения к Redis
	options := redis.Options{
		Address:   fmt.Sprintf("%s:%d", cfg.HostName, cfg.Port),
		Password:  cfg.Password,
		MaxMemory: "100mb",
		Policy:    "allkeys-lru",
	}

	// пробуем подключиться
	client, err := redis.Connect(options)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка установки соединения с Redis: %v\n", err)
	}

	// проверяем подключение
	if err := client.Ping(context.Background()); err != nil {
		return nil, nil, fmt.Errorf("ошибка подключения к Redis: %v\n", err)
	}

	log.Info("Рэдис подключен.")

	// оборачиваем клиент в структуру с методами ZSET
	clientRedis := &ClientZSet{Client: client}

	// канал для передачи номеров просроченных броней брокеру
	ch := make(chan int, 100)

	// запускаем горутину-наблюдатель
	go watchOverdueLoop(ctx, clientRedis, cfg, ch, log)

	return clientRedis, ch, nil
}
