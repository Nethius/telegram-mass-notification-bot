package main

import (
	"configuration_parser/internal/command_parser"
	"configuration_parser/internal/repository/postgres"
	"configuration_parser/internal/telegram_api"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
)

func main() {
	mainLogger := initLogger()
	logger := mainLogger.With().Str("component", "Main").Logger()

	err := godotenv.Load()
	if err != nil {
		logger.Panic().Msg(err.Error())
	}

	connStr, err := getPostgresCredentials()
	if err != nil {
		logger.Panic().Msgf("failed to get db credentials from env: %v", err)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Panic().Msgf("failed to open db connection: %v", err)
	}
	defer db.Close()

	repository := postgres.NewRepository(db)
	parser := command_parser.NewService(logger, repository)
	tgApiService := telegram_api.NewService(logger, *parser)

	logger.Fatal().Msgf("failed to listen telegram api server: %v", tgApiService.ListenAndServe())
}

func initLogger() zerolog.Logger {
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	return zerolog.New(os.Stdout).With().Timestamp().Logger()
}

func getPostgresCredentials() (string, error) {
	host, ok := os.LookupEnv("PGHOST")
	if !ok {
		return "", errors.New("failed to get PGHOST from env")
	}

	port, ok := os.LookupEnv("PGPORT")
	if !ok {
		return "", errors.New("failed to get PGPORT from env")
	}

	user, ok := os.LookupEnv("PGUSER")
	if !ok {
		return "", errors.New("failed to get PGUSER from env")
	}

	password, ok := os.LookupEnv("PGPASSWORD")
	if !ok {
		return "", errors.New("failed to get PGPASSWORD from env")
	}

	dbname, ok := os.LookupEnv("PGDATABASE")
	if !ok {
		return "", errors.New("failed to get PGDATABASE from env")
	}

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname), nil
}
