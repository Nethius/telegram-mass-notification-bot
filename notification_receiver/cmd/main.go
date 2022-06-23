package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"notification_receiver/internal/repository/postgres"
	"os"

	"notification_receiver/internal/handlers"
	"notification_receiver/internal/model"
	"notification_receiver/internal/publisher"

	"github.com/gorilla/mux"
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

	notifications := make(chan model.Notification)
	addNotificationHandler := addNotifications.NewHandler(repository, logger, notifications)

	publisherService := publisher.NewService(logger, notifications)

	// TODO handle errors
	go publisherService.StartPublishing()

	httpServerCredentials, err := getHttpServerCredentials()
	if err != nil {
		logger.Panic().Msgf("failed to get http server credentials from env: %v", err)
	}

	router := mux.NewRouter()
	// TODO change api endpoint
	router.HandleFunc("/", addNotificationHandler.AddNotification).Methods("POST")

	logger.Fatal().Msgf("failed to listen ws server: %v", http.ListenAndServe(httpServerCredentials, router))
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
