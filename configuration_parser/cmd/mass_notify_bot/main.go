package main

import (
	"configuration_parser/internal/command_parser"
	"configuration_parser/internal/repository/postgres"
	"database/sql"
	"errors"
	"fmt"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
	parser := command_parser.NewService(repository)

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		logger.Panic().Msg(err.Error())
	}

	bot.Debug = true //TODO remove

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		//TODO логгировать все сообщения ?

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		switch update.Message.Command() {
		case "start":
			msg.Text = parser.Start(update.Message.Chat.ID, update.Message.Chat.UserName)
		case "grant_access":
			msg.Text = parser.GrantAccess(update.Message.Chat.ID, update.Message.Text)
		case "remove_access":
			msg.Text = parser.RemoveAccess(update.Message.Chat.ID, update.Message.Text)
		default:
			msg.Text = "Command list:\n\n" +
				"/start - join the list of active users.\n\n" +
				"/exit - ?\n\n" +
				"/grant_access @username - let user - @username send me notifications.\n\n" +
				"/remove_access @username - prevent user - @username send me notifications."
		}

		if _, err := bot.Send(msg); err != nil {
			logger.Panic().Msg(err.Error()) // TODO not panic
		}
	}
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
