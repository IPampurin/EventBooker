package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/IPampurin/EventBooker/pkg/configuration"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/logger"
	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/retry"
)

// Broker содержит Publisher для отправки сообщений и канал для получения просроченных броней.
type Broker struct {
	Publisher *rabbitmq.Publisher
	Messages  <-chan int // канал, из которого сервисный слой читает ID просроченных броней
	closeFunc func() error
}

// InitMQ создаёт подключение к RabbitMQ, объявляет exchange/очередь,
// запускает consumer и возвращает Broker.
// consumer автоматически пишет ID броней (int) в канал Messages.
// При ошибке или закрытии клиента канал закрывается.
func InitMQ(ctx context.Context, cfg *configuration.ConfBroker, log logger.Logger) (*Broker, error) {

	// Формируем URL
	amqpURL := fmt.Sprintf("amqp://%s:%s@%s:%d%s",
		cfg.User, cfg.Password, cfg.HostName, cfg.Port, cfg.VHost)

	// Единая стратегия повторов
	retryStrategy := retry.Strategy{
		Attempts: cfg.RetryAttempts,
		Delay:    cfg.RetryDelay,
		Backoff:  cfg.RetryBackoff,
	}

	// Конфигурация клиента
	clientCfg := rabbitmq.ClientConfig{
		URL:            amqpURL,
		ConnectionName: cfg.ConnName,
		ConnectTimeout: cfg.ConnTimeout,
		Heartbeat:      cfg.Heartbeat,
		ReconnectStrat: retryStrategy,
		ProducingStrat: retryStrategy,
		ConsumingStrat: retryStrategy,
	}

	// Создаём клиента
	client, err := rabbitmq.NewClient(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq.NewClient: %w", err)
	}

	// Объявляем exchange (durable = true)
	if err = client.DeclareExchange(cfg.Exchange, "direct", true, false, false, nil); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("DeclareExchange: %w", err)
	}

	// Объявляем очередь и привязываем к exchange
	if err = client.DeclareQueue(cfg.Queue, cfg.Exchange, cfg.RoutingKey, true, false, true, nil); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("DeclareQueue: %w", err)
	}

	// Создаём Publisher
	publisher := rabbitmq.NewPublisher(client, cfg.Exchange, "application/json")

	// Создаём канал для передачи ID броней (буферизированный, чтобы не блокировать consumer)
	msgCh := make(chan int, 100)

	// Определяем обработчик сообщений consumer'а
	handler := func(ctx context.Context, d amqp.Delivery) error {
		// Ожидаем, что тело сообщения содержит ID брони (как число в JSON или просто число)
		var id int
		// Пробуем распарсить как JSON-число
		if err := json.Unmarshal(d.Body, &id); err != nil {
			// Если не JSON, пробуем преобразовать строку в int
			id, err = strconv.Atoi(string(d.Body))
			if err != nil {
				log.Error("invalid message body", "body", string(d.Body))
				// Отбрасываем сообщение без повторной постановки в очередь
				_ = d.Nack(false, false) // multiple=false, requeue=false
				return fmt.Errorf("invalid body: %w", err)
			}
		}

		// Пытаемся отправить ID в канал
		select {
		case msgCh <- id:
			// Успешно отправили, подтверждаем сообщение
			return nil // consumer выполнит ack
		case <-ctx.Done():
			// Контекст отменён – не подтверждаем, но и не паникуем
			_ = d.Nack(false, true) // requeue, чтобы сообщение не потерялось
			return ctx.Err()
		}
	}

	// Конфигурация consumer'а (параметры можно будет вынести в конфиг при необходимости)
	consumerCfg := rabbitmq.ConsumerConfig{
		Queue:         cfg.Queue,
		AutoAck:       false,
		Ask:           rabbitmq.AskConfig{Multiple: false},
		Nack:          rabbitmq.NackConfig{Multiple: false, Requeue: true},
		Workers:       1,
		PrefetchCount: 1,
	}
	consumer := rabbitmq.NewConsumer(client, consumerCfg, handler)

	// Запускаем consumer в горутине
	go func() {
		defer close(msgCh) // при завершении consumer закрываем канал
		if err := consumer.Start(ctx); err != nil {
			log.Error("rabbitmq consumer stopped", "error", err)
		}
	}()

	broker := &Broker{
		Publisher: publisher,
		Messages:  msgCh,
		closeFunc: client.Close,
	}

	log.Info("RabbitMQ инициализирован", "exchange", cfg.Exchange, "queue", cfg.Queue)
	return broker, nil
}

// Close закрывает соединение с RabbitMQ (останавливает продюсер и консумер)
func (b *Broker) Close() error {

	if b.closeFunc != nil {
		return b.closeFunc()
	}

	return nil
}
