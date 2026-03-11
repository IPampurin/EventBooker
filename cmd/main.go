package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/IPampurin/EventBooker/pkg/config"
	"github.com/IPampurin/EventBooker/pkg/db"
	"github.com/IPampurin/EventBooker/pkg/server"
	"github.com/wb-go/wbf/logger"
)

func main() {

	// cоздаём контекст
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// запускаем горутину обработки сигналов
	go signalHandler(ctx, cancel)

	// считываем .env файл
	cfg, err := config.ReadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// настраиваем логгер
	appLogger, err := logger.InitLogger(
		logger.ZapEngine,
		"EventBooker",
		os.Getenv("APP_ENV"), // пока оставим пустым
		logger.WithLevel(logger.InfoLevel),
	)
	if err != nil {
		log.Fatalf("Ошибка создания логгера: %v", err)
	}
	defer func() { _ = appLogger.(*logger.ZapAdapter) }()

	// получаем экземпляр БД
	storageDB, err := db.InitDB(ctx, &cfg.DB, appLogger)
	if err != nil {
		appLogger.Error("ошибка подключения к БД", "error", err)
		return
	}
	defer func() { _ = db.CloseDB(storageDB) }()

	// получаем broker,error (структура паблишер RabbitMQ и консумер RabbitMQ)

	// получаем экземпляр zSet, error (Redis просто реализует положить-отдать для использования методов в сервисном слое)

	// заводим канал transferOverBookings для передачи номеров броней (из горутины получения просроченных броней из RabbitMQ в сервисный слой для отмены брони)

	// получаем экземпляр слоя бизнес-логики (передаём storageDB, zSet, broker, transferOverBookings)

	// запускаем горутину получения из Redis просроченных броней (передаём паблишер RabbitMQ)

	// запускаем горутину получения просроченных броней из RabbitMQ (передаём консумер RabbitMQ)

	// запускаем сервер
	err = server.Run(ctx, &cfg.Server, service, appLogger)
	if err != nil {
		appLogger.Error("Ошибка сервера", "error", err)
		cancel()
		return
	}

	appLogger.Info("Приложение корректно завершено")
}

// signalHandler обрабатывет сигналы отмены
func signalHandler(ctx context.Context, cancel context.CancelFunc) {

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	select {
	case <-ctx.Done():
		return
	case <-sigChan:
		cancel()
		return
	}
}
