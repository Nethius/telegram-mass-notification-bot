package publisher

import (
	"encoding/json"

	"notification_receiver/internal/model"

	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
)

type Service struct {
	logger        zerolog.Logger
	notifications chan model.Notification
}

func NewService(logger zerolog.Logger, notifications chan model.Notification) *Service {
	l := logger.With().Str("component", "publisher").Logger()

	return &Service{
		logger:        l,
		notifications: notifications,
	}
}

func (s *Service) StartPublishing() error {
	// todo add parse cred
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return err
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	queue, err := channel.QueueDeclare(
		"notification_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	for {
		notification := <-s.notifications
		message, err := json.Marshal(notification)
		if err != nil {
			s.logger.Error().Msgf("Error while encode new notification, error: %s", err.Error())
		}

		err = channel.Publish(
			"",
			queue.Name,
			false,
			false,
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        message,
			})
		if err != nil {
			s.logger.Error().Msgf("Error while publish new notification to message broker, error: %s", err.Error())
		}
	}
}
