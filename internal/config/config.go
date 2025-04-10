package config

import (
	"time"
)

// Общая конфигурация сервиса, тут должны быть все переменные

type AppConfig struct {
	LogLevel   string
	Rest       Rest
}

type Rest struct {
	ListenAddress string        `envconfig:"PORT" required:"true"`
	WriteTimeout  time.Duration `envconfig:"WRITE_TIMEOUT" required:"true"`
	ServerName    string        `envconfig:"SERVER_NAME" required:"true"`
	Token         string        `envconfig:"TOKEN" required:"true"`
}
