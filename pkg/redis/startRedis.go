package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/IPampurin/EventBooker/pkg/configuration"

	"github.com/wb-go/wbf/logger"
	"github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/retry"

	goredis "github.com/go-redis/redis/v8" // оригинальный go-redis для типов ZSET
)

// ClientRedis хранит подключение к БД Redis
type ClientRedis struct {
	*redis.Client
}

// ZRangeByScore получает элементы из сортированного множества с баллами в интервале [min, max]
func (c *ClientRedis) ZRangeByScore(ctx context.Context, key string, min, max int64) ([]string, error) {

	// используем оригинальный тип goredis.ZRangeBy
	return c.Client.ZRangeByScore(ctx, key, &goredis.ZRangeBy{
		Min: strconv.FormatInt(min, 10),
		Max: strconv.FormatInt(max, 10),
	}).Result()
}

// ZRem удаляет элемент из сортированного множества
func (c *ClientRedis) ZRem(ctx context.Context, key string, members ...interface{}) error {

	return c.Client.ZRem(ctx, key, members...).Err()
}

// InitRedis запускает работу с Redis и горутину для отслеживания просроченных бронирований
func InitRedis(ctx context.Context, cfg *configuration.ConfZSet, log logger.Logger) (<-chan int, error) {

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
		return nil, fmt.Errorf("ошибка установки соединения с Redis: %v\n", err)
	}

	// проверяем подключение
	if err := client.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ошибка подключения к Redis: %v\n", err)
	}

	log.Info("Рэдис подключен.")

	// оборачиваем клиент в структуру с методами ZSET
	clientRedis := &ClientRedis{Client: client}

	// канал для передачи номеров просроченных броней брокеру
	ch := make(chan int, 100)

	// запускаем горутину-наблюдатель
	go watchOverdueLoop(ctx, clientRedis, cfg, ch, log)

	return ch, nil
}

// watchOverdueLoop периодически проверяет ZSET и отправляет найденные ID в канал
func watchOverdueLoop(ctx context.Context, client *ClientRedis, cfg *configuration.ConfZSet, ch chan<- int, log logger.Logger) {

	defer close(ch)      // закрываем канал при выходе
	defer client.Close() // закрываем соединение с Redis

	ticker := time.NewTicker(cfg.CheckInterval)
	defer ticker.Stop()

	// стратегия повторов для чтения из Redis
	readRetry := retry.Strategy{
		Attempts: cfg.ReadRetryAttempts,
		Delay:    cfg.ReadRetryDelay,
		Backoff:  cfg.ReadRetryBackoff,
	}

	for {
		select {
		case <-ctx.Done():
			log.Info("горутина наблюдателя Redis остановлена по контексту")
			return

		case <-ticker.C:
			now := time.Now().Unix()

			// читаем из ZSET с повторными попытками
			var members []string
			err := retry.DoContext(ctx, readRetry, func() error {
				var e error
				members, e = client.ZRangeByScore(ctx, cfg.OverdueKey, 0, now)
				if e != nil && e != redis.NoMatches {
					return e // любая ошибка, кроме "нет элементов", запустит повтор
				}
				return nil // успех или redis.NoMatches (нет элементов)
			})

			if err != nil {
				log.Error("ошибка чтения ZSET после всех попыток", "error", err)
				continue
			}

			// обрабатываем каждый найденный элемент
			for _, member := range members {

				id, err := strconv.Atoi(member)
				if err != nil {
					log.Error("неверный формат ID в Redis", "value", member)
					continue
				}

				select {
				case ch <- id:
					// успешно отправили – удаляем элемент из ZSET
					if err := client.ZRem(ctx, cfg.OverdueKey, member); err != nil {
						log.Error("не удалось удалить элемент из ZSET", "id", id, "error", err)
					} else {
						log.Info("просроченная бронь отправлена и удалена из Redis", "id", id)
					}
				case <-ctx.Done():
					return
				}
			}
		}
	}
}
