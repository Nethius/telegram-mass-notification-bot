package telegram_api

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
	"os"
)

type parser interface {
	Start(userId int64, userName string) string
	GrantAccess(userId int64, request string) string
	RemoveAccess(userId int64, request string) string
}

type Service struct {
	logger zerolog.Logger
	parser parser
}

func NewService(logger zerolog.Logger, commandParser parser) *Service {
	l := logger.With().Str("component", "telegram_api").Logger()

	return &Service{
		logger: l,
		parser: commandParser,
	}
}

func (s *Service) ListenAndServe() error {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		return err
	}

	// TODO change offset from 0? change timeout?
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		s.logger.Info().Msgf("received message from tg api. id: %v, nickname: %v, message: %v",
			update.Message.Chat.ID, update.Message.Chat.UserName, update.Message.Text)

		go func(update tgbotapi.Update) {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

			switch update.Message.Command() {
			//TODO add command to view list of all user with access
			case "start":
				msg.Text = s.parser.Start(update.Message.Chat.ID, update.Message.Chat.UserName)
			case "grant_access":
				msg.Text = s.parser.GrantAccess(update.Message.Chat.ID, update.Message.Text)
			case "remove_access":
				msg.Text = s.parser.RemoveAccess(update.Message.Chat.ID, update.Message.Text)
			default:
				msg.Text = "Command list:\n\n" +
					"/start - join the list of active users.\n\n" +
					"/exit - ?\n\n" +
					"/grant_access @username - let user - @username send me notifications.\n\n" +
					"/remove_access @username - prevent user - @username send me notifications."
			}

			if _, err := bot.Send(msg); err != nil {
				s.logger.Error().Msgf("failed to send response to telegram, %v", err)
			} else {
				s.logger.Info().Msgf("send message to tg api. id: %v, nickname: %v, message: %v",
					update.Message.Chat.ID, update.Message.Chat.UserName, msg.Text)
			}
		}(update)
	}
	return ErrUnexpected
}
