package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"notification_receiver/internal/repository/postgres"
	"os"

	"notification_receiver/internal/handlers"
	"notification_receiver/internal/publisher"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
)

func main() {
	mainLogger := initLogger()
	logger := mainLogger.With().Str("component", "main").Logger()

	err := godotenv.Load()
	if err != nil {
		logger.Panic().Msg(err.Error())
	}

	postgresCredentials, err := getPostgresCredentials()
	if err != nil {
		logger.Panic().Msgf("failed to get db credentials from env: %v", err)
	}

	db, err := sql.Open("postgres", postgresCredentials)
	if err != nil {
		logger.Panic().Msgf("failed to open db connection: %v", err)
	}
	// TODO handle error
	defer db.Close()

	repository := postgres.NewRepository(db)

	publisherService, err := publisher.NewService(logger)
	if err != nil {
		logger.Panic().Msgf("failed to connect to rabbitmq: %v", err)
	}
	// TODO handle error
	defer publisherService.Close()

	addNotificationHandler := addNotifications.NewHandler(repository, logger, publisherService)

	httpServerCredentials, err := getHttpServerCredentials()
	if err != nil {
		logger.Panic().Msgf("failed to get http server credentials from env: %v", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/add-notification", addNotificationHandler.AddNotification).Methods("POST")

	logger.Fatal().Msgf("failed to listen http server: %v", http.ListenAndServe(httpServerCredentials, router))
}

func initLogger() zerolog.Logger {
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	return zerolog.New(os.Stdout).With().Timestamp().Logger()
}

func getHttpServerCredentials() (string, error) {
	host, ok := os.LookupEnv("HTTPSERVERHOST")
	if !ok {
		return "", errors.New("failed to get HTTPSERVERHOST from env")
	}

	port, ok := os.LookupEnv("HTTPSERVERPORT")
	if !ok {
		return "", errors.New("failed to get HTTPSERVERPORT from env")
	}

	return fmt.Sprintf("%s:%s", host, port), nil
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
