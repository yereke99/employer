package logger

import (
	"os"

	"go.uber.org/zap"
)

// NewLogger создает новый logger
func NewLogger() (*zap.Logger, error) {
	env := os.Getenv("ENVIRONMENT")

	var config zap.Config

	if env == "production" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	// Настройка кодировщика для более читаемого вывода
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.MessageKey = "message"

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}
