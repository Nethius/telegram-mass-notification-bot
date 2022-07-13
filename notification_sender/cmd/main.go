package main

import (
	"os"

	"notification_sender/internal/consumer"
	"notification_sender/internal/sender"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

func main() {
	mainLogger := initLogger()
	logger := mainLogger.With().Str("component", "main").Logger()

	err := godotenv.Load()
	if err != nil {
		logger.Panic().Msg(err.Error())
	}

	sendingService, err := sender.NewService(logger)
	if err != nil {
		logger.Panic().Msgf("failed to connect to telegram api: %v", err)
	}

	consumerService, err := consumer.NewService(logger, sendingService)
	if err != nil {
		logger.Panic().Msgf("failed to connect to rabbitmq: %v", err)
	}
	// TODO handle errors
	defer consumerService.Close()

	// TODO handle errors
	consumerService.StartConsuming()

}

func initLogger() zerolog.Logger {
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	return zerolog.New(os.Stdout).With().Timestamp().Logger()
}
