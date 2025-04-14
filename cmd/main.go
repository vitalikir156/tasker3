package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"

	"github.com/vitalikir156/tasker3/internal/api"
	"github.com/vitalikir156/tasker3/internal/config"
	customLogger "github.com/vitalikir156/tasker3/internal/logger"
	"github.com/vitalikir156/tasker3/internal/repo"
	"github.com/vitalikir156/tasker3/internal/service"

	"github.com/joho/godotenv"
)

func main() {

	loadenv := pflag.BoolP("loadenv", "e", false, "load .env file")
	pflag.Parse()

	if *loadenv {
		err := godotenv.Load()
		if err != nil {
			log.Fatal(errors.Wrap(err, "failed to load env file"))
		}
	}

	// Загружаем конфигурацию из переменных окружения
	var cfg config.AppConfig
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(errors.Wrap(err, "failed to load configuration"))
	}

	// Инициализация логгера
	logger, err := customLogger.NewLogger(cfg.LogLevel)
	if err != nil {
		log.Fatal(errors.Wrap(err, "error initializing logger"))
	}


	repository, err := repo.NewRepository(context.Background(), cfg.PostgreSQL)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to initialize repository"))
	}

	// Создание сервиса с бизнес-логикой
	serviceInstance := service.NewService(repository, logger)

	// Инициализация API
	app := api.NewRouters(&api.Routers{Service: serviceInstance}, cfg.Rest.Token)

	// Запуск HTTP-сервера в отдельной горутине
	go func() {
		logger.Infof("Starting server on %s", cfg.Rest.ListenAddress)
		if err := app.Listen(cfg.Rest.ListenAddress); err != nil {
			log.Fatal(errors.Wrap(err, "failed to start server"))
		}
	}()

	// Ожидание системных сигналов для корректного завершения работы
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	<-signalChan

	logger.Info("Shutting down gracefully...")
}
