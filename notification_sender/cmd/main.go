package main

import (
	"os"

	"notification_sender/internal/consumer"
	"notification_sender/internal/model"
	"notification_sender/internal/sender"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

func main() {
	mainLogger := initLogger()
	logger := mainLogger.With().Str("component", "Main").Logger()

	err := godotenv.Load()
	if err != nil {
		logger.Panic().Msg(err.Error())
	}

	notifications := make(chan model.Notification)
	consumerService := consumer.NewService(logger, notifications)

	// TODO handle errors
	go consumerService.StartConsuming()

	sendingService := sender.NewService(logger, notifications)
	// TODO handle errors
	sendingService.StartSending()
}

func initLogger() zerolog.Logger {
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	return zerolog.New(os.Stdout).With().Timestamp().Logger()
}
