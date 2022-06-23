package consumer

import (
	"encoding/json"
	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
	"notification_sender/internal/model"
)

type Service struct {
	logger        zerolog.Logger
	notifications chan model.Notification
}

func NewService(logger zerolog.Logger, notifications chan model.Notification) *Service {
	l := logger.With().Str("component", "consumer").Logger()

	return &Service{
		logger:        l,
		notifications: notifications,
	}
}

func (s *Service) StartConsuming() error {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
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

	err = ch.Qos(
		1,
		0,
		false,
	)
	if err != nil {
		return err
	}

	messages, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	for message := range messages {
		s.logger.Info().Msgf("Received a message from message publisher: %s", message.Body)
		var notification model.Notification
		err := json.Unmarshal(message.Body, &notification)
		if err != nil {
			s.logger.Error().Msgf("failed to decode request %v: %v", message.Body, err)
		}
		s.notifications <- notification
		err = message.Ack(false)
		if err != nil {
			s.logger.Error().Msgf("failed to send ack message to rmq")
		}

	}
	return nil
}
