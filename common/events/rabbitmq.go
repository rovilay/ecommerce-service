package events

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitClient struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

// func ConnectRabbit(username, password, host string, port uint16) (*amqp.Connection, error) {
func ConnectRabbit(url string) (*amqp.Connection, error) {
	// url := fmt.Sprintf("amqp://%s:%s@%s:%d", username, password, host, port)
	conn, err := amqp.Dial(url)

	return conn, err
}

func NewRabbitClient(conn *amqp.Connection, topic Topic) (*RabbitClient, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	rc := &RabbitClient{
		conn: conn,
		ch:   ch,
	}

	// create a durable exchange
	err = rc.createExchange(topic, true, false)
	if err != nil {
		return nil, err
	}

	return rc, nil
}

// create exchange
func (rc *RabbitClient) createExchange(topic Topic, durable, autodelete bool) error {
	return rc.ch.ExchangeDeclare(string(topic), "topic", durable, autodelete, false, false, nil)
}

// create queue
func (rc *RabbitClient) CreateQueue(queueName string, durable, autodelete bool) (amqp.Queue, error) {
	return rc.ch.QueueDeclare(queueName, durable, autodelete, false, false, nil)
}

// CreateBinding is used to connect a queue to an Exchange using the binding rule
func (rc *RabbitClient) CreateBinding(exchange Topic, queueName string, bindingKey RoutingKey) error {
	// leaving nowait false, having nowait set to false will cause the channel to return an error and close if it cannot bind
	return rc.ch.QueueBind(queueName, string(bindingKey), string(exchange), false, nil)
}

func (rc *RabbitClient) Send(ctx context.Context, exchange, routingKey string, event EventData) error {
	b, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return rc.ch.PublishWithContext(ctx,
		exchange,   // exchange
		routingKey, // routing key
		// Mandatory is used when we HAVE to have the message return an error, if there is no route or queue then
		// setting this to true will make the message bounce back
		// If this is False, and the message fails to deliver, it will be dropped
		true, // mandatory
		// immediate Removed in MQ 3 or up https://blog.rabbitmq.com/posts/2012/11/breaking-things-with-rabbitmq-3-0§
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        b,
		},
	)
}

// Consume is a wrapper around consume, it will return a Channel that can be used to digest messages
// Queue is the name of the queue to Consume
// Consumer is a unique identifier for the service instance that is consuming, can be used to cancel etc
// autoAck is important to understand, if set to true, it will automatically Acknowledge that processing is done
// This is good, but remember that if the Process fails before completion, then an ACK is already sent, making a message lost
// if not handled properly
func (rc *RabbitClient) Consume(queue RoutingKey, consumer Topic, autoAck bool) (<-chan amqp.Delivery, error) {
	return rc.ch.Consume(string(queue), string(consumer), autoAck, false, false, false, nil)
}

// close the channel
func (rc *RabbitClient) Close() error {
	return rc.ch.Close()
}