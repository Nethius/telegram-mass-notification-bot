package main

import (
	"configuration_parser/internal/command_parser"
	"configuration_parser/internal/repository/postgres"
	"configuration_parser/internal/telegram_api"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
)

func main() {
	mainLogger := initLogger()
	logger := mainLogger.With().Str("component", "main").Logger()

	err := godotenv.Load()
	if err != nil {
		logger.Panic().Msgf("failed to parse env: %v", err)
	}

	repository, err := postgres.NewRepository()
	if err != nil {
		logger.Panic().Msgf("failed to setup repository: %v", err)
	}
	defer repository.Close()

	parser := command_parser.NewService(repository)
	tgApiService, err := telegram_api.NewService(logger, parser)
	if err != nil {
		logger.Panic().Msgf("failed to setup telegram api: %v", err)
	}

	logger.Fatal().Msgf("failed to listen telegram api server: %v", tgApiService.ListenAndServe())
}

func initLogger() zerolog.Logger {
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	return zerolog.New(os.Stdout).With().Timestamp().Logger()
}
