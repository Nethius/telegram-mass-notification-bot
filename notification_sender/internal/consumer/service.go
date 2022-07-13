package consumer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
	"notification_sender/internal/model"
	"os"
)

type sender interface {
	Send(notification model.Notification) error
}

type Service struct {
	logger        zerolog.Logger
	sender        sender
	rmqConnection *amqp.Connection
	rmqChannel    *amqp.Channel
	rmqQueue      *amqp.Queue
}

func NewService(logger zerolog.Logger, sender sender) (*Service, error) {
	l := logger.With().Str("component", "consumer").Logger()

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

	err = channel.Qos(
		1,
		0,
		false,
	)
	if err != nil {
		return nil, err
	}

	return &Service{
		logger:        l,
		sender:        sender,
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

func (s *Service) StartConsuming() error {
	messages, err := s.rmqChannel.Consume(
		s.rmqQueue.Name,
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
		s.logger.Info().Msgf("received a message from message broker: %s", message.Body)
		var notification model.Notification
		err := json.Unmarshal(message.Body, &notification)
		if err != nil {
			s.logger.Error().Msgf("failed to decode message %v: %v", message.Body, err)
		}

		err = s.sender.Send(notification)
		if err != nil {
			err = message.Nack(false, true)
		} else {
			err = message.Ack(false)
		}
		if err != nil {
			s.logger.Error().Msgf("failed to send response to message broker")
		}
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
