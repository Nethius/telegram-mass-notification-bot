package sender

import (
	"os"
	"strconv"

	"notification_sender/internal/model"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
)

type Service struct {
	logger zerolog.Logger
	botApi *tgbotapi.BotAPI
}

func NewService(logger zerolog.Logger) (*Service, error) {
	l := logger.With().Str("component", "sender").Logger()

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		return nil, err
	}

	return &Service{
		logger: l,
		botApi: bot,
	}, nil
}

func (s *Service) Send(notification model.Notification) error {
	for _, recipient := range notification.RecipientsId {
		id, _ := strconv.ParseInt(recipient, 10, 64)
		message := tgbotapi.NewMessage(id, notification.Message)

		if _, err := s.botApi.Send(message); err != nil {
			s.logger.Error().Msgf("failed to send message to telegram: %v", err)
			return err
		}
	}
	return nil
}
