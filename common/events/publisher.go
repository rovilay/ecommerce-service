package events

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

type Publisher interface {
	Publish(ctx context.Context, routingkey RoutingKey, data string) error
}

type eventPublisher struct {
	topic Topic
	log   *zerolog.Logger
	ch    *amqp.Channel
}

func NewEventPublisher(ch *amqp.Channel, topic Topic, log *zerolog.Logger) (*eventPublisher, error) {
	eventLogger := log.With().Str("publisher", string(topic)).Logger()

	// create the exchange
	err := ch.ExchangeDeclare(
		string(topic),
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		eventLogger.Err(err).Msg("failed to declair an exchage")
		return nil, err
	}

	publisher := &eventPublisher{
		topic: topic,
		ch:    ch,
		log:   &eventLogger,
	}

	return publisher, err
}

func (p *eventPublisher) Publish(ctx context.Context, routingkey RoutingKey, event EventData) error {
	body, err := json.Marshal(event)
	if err != nil {
		p.log.Err(err).Msg("failed to marshal event data")

		return err
	}

	err = p.ch.PublishWithContext(ctx,
		string(p.topic), // exchange
		string(routingkey),
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})

	return err
}
