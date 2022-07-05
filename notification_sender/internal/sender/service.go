package sender

import (
	"os"
	"strconv"

	"notification_sender/internal/model"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
)

type Service struct {
	logger        zerolog.Logger
	notifications chan model.Notification
}

func NewService(logger zerolog.Logger, notifications chan model.Notification) *Service {
	l := logger.With().Str("component", "sender").Logger()

	return &Service{
		logger:        l,
		notifications: notifications,
	}
}

func (s *Service) StartSending() error {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		return err
	}

	for {
		notification := <-s.notifications

		for _, recipient := range notification.RecipientsId {
			s.logger.Info().Msgf("received a message from message consumer: %s", recipient)
			id, _ := strconv.ParseInt(recipient, 10, 64)
			message := tgbotapi.NewMessage(id, notification.Message)

			if _, err := bot.Send(message); err != nil {
				s.logger.Error().Msgf("failed to send message to telegram: %v", err)
			}
		}
	}
}
