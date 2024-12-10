package events

import (
	"context"
	"encoding/json"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

type Listener interface {
	Lonsume(ctx context.Context)
}

type eventListener struct {
	topic Topic
	log   *zerolog.Logger
	ch    *amqp.Channel
	q     *amqp.Queue
}

func NewEventListener(ch *amqp.Channel, topic Topic, routingkeys []RoutingKey, log *zerolog.Logger) (*eventListener, error) {
	eventLogger := log.With().Str("listener", string(topic)).Logger()

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
		return nil, err
	}

	// create the queue
	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		return nil, err
	}

	for _, s := range routingkeys {
		err = ch.QueueBind(
			q.Name,        // queue name
			string(s),     // routing key
			string(topic), // exchange
			false,
			nil,
		)

		if err != nil {
			return nil, err
		}
	}

	return &eventListener{
		log:   &eventLogger,
		topic: topic,
		ch:    ch,
		q:     &q,
	}, nil
}

func (l *eventListener) Listen(ctx context.Context, msgHandler func(e EventData, msg amqp.Delivery)) error {
	msgs, err := l.ch.ConsumeWithContext(ctx,
		l.q.Name, // queue
		"",       // consumer
		true,     // auto ack
		false,    // exclusive
		false,    // no local
		false,    // no wait
		nil,      // args
	)

	if err != nil {
		l.log.Err(err).Msg("failed to consume messages")
		return err
	}

	var forever chan struct{}

	go func() {
		wg := sync.WaitGroup{}
		for msg := range msgs {
			wg.Add(1)

			go func() {
				defer wg.Done()

				e := EventData{}
				if err = json.Unmarshal(msg.Body, &e); err != nil {
					l.log.Err(err).Msg("failed to unmarshal event")
					return
				}

				msgHandler(e, msg)
			}()

			wg.Wait()
		}
	}()

	<-forever
	return nil
}
