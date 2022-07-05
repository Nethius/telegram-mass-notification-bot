package publisher

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"notification_receiver/internal/model"

	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
)

type Service struct {
	logger        zerolog.Logger
	rmqConnection *amqp.Connection
	rmqChannel    *amqp.Channel
	rmqQueue      *amqp.Queue
}

func NewService(logger zerolog.Logger) (*Service, error) {
	l := logger.With().Str("component", "publisher").Logger()

	rabbitMqCredentials, err := getRabbitMqCredentials()
	if err != nil {
		return nil, err
	}
	conn, err := amqp.Dial(rabbitMqCredentials)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	queue, err := channel.QueueDeclare(
		"notification_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &Service{
		logger:        l,
		rmqConnection: conn,
		rmqChannel:    channel,
		rmqQueue:      &queue,
	}, nil
}

func (s *Service) Close() error {
	// TODO if the channel is closed with an error, the connection will not be closed?
	err := s.rmqChannel.Close()
	if err != nil {
		return err
	}
	err = s.rmqConnection.Close()
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) Publish(notification model.Notification) error {
	message, err := json.Marshal(notification)
	if err != nil {
		s.logger.Error().Msgf("error while encode new notification, error: %s", err.Error())
		return ErrInternal
	}

	err = s.rmqChannel.Publish(
		"",
		s.rmqQueue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        message,
		})
	if err != nil {
		s.logger.Error().Msgf("error while publish new notification to message broker, error: %s", err.Error())
		return ErrInternal
	}
	return nil
}

func getRabbitMqCredentials() (string, error) {
	username, ok := os.LookupEnv("RABBITMQ_USERNAME")
	if !ok {
		return "", errors.New("failed to get RABBITMQ_USERNAME from env")
	}

	password, ok := os.LookupEnv("RABBITMQ_PASSWORD")
	if !ok {
		return "", errors.New("failed to get RABBITMQ_PASSWORD from env")
	}

	host, ok := os.LookupEnv("RABBITMQ_HOST")
	if !ok {
		return "", errors.New("failed to get RABBITMQ_HOST from env")
	}

	port, ok := os.LookupEnv("RABBITMQ_PORT")
	if !ok {
		return "", errors.New("failed to get RABBITMQ_PORT from env")
	}

	return fmt.Sprintf("amqp://%s:%s@%s:%s", username, password, host, port), nil
}
