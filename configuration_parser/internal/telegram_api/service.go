package telegram_api

import (
	"errors"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
)

type parser interface {
	Start(userId int64, userName string) (string, error)
	GrantAccess(userId int64, request string) (string, error)
	RemoveAccess(userId int64, request string) (string, error)
}

type Service struct {
	logger zerolog.Logger
	parser parser
	bot    *tgbotapi.BotAPI
}

func NewService(logger zerolog.Logger, commandParser parser) (*Service, error) {
	l := logger.With().Str("component", "telegram_api").Logger()

	token, ok := os.LookupEnv("TELEGRAM_APITOKEN")
	if !ok {
		return nil, errors.New("failed to get TELEGRAM_APITOKEN from env")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &Service{
		logger: l,
		parser: commandParser,
		bot:    bot,
	}, nil
}

func (s *Service) ListenAndServe() error {
	// TODO change offset from 0? change timeout?
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	updates := s.bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		s.logger.Info().Msgf("received message from tg api. id: %v, nickname: %v, message: %v",
			update.Message.Chat.ID, update.Message.Chat.UserName, update.Message.Text)

		go s.handleMessage(update)
	}
	return ErrUnexpected
}

func (s *Service) handleMessage(update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

	var err error
	switch update.Message.Command() {
	//TODO add command to view list of all user with access
	case "start":
		msg.Text, err = s.parser.Start(update.Message.Chat.ID, update.Message.Chat.UserName)
	case "grant_access":
		msg.Text, err = s.parser.GrantAccess(update.Message.Chat.ID, update.Message.Text)
	case "remove_access":
		msg.Text, err = s.parser.RemoveAccess(update.Message.Chat.ID, update.Message.Text)
	default:
		msg.Text = "Command list:\n\n" +
			"/start - join the list of active users.\n\n" +
			"/exit - ?\n\n" +
			"/grant_access @username - let user - @username send me notifications.\n\n" +
			"/remove_access @username - prevent user - @username send me notifications."
	}

	if err != nil {
		s.logger.Error().Msgf("error while process command, %v", err)
	}

	if _, err = s.bot.Send(msg); err != nil {
		s.logger.Error().Msgf("failed to send response to telegram, %v", err)
	} else {
		s.logger.Info().Msgf("send message to tg api. id: %v, nickname: %v, message: %v",
			update.Message.Chat.ID, update.Message.Chat.UserName, msg.Text)
	}
}
