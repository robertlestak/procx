package client

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	RabbitMQClient *amqp.Connection
	RabbitMQQueue  string
)

func CreateRabbitMQClient(url string) error {
	conn, err := amqp.Dial(url)
	if err != nil {
		return err
	}
	RabbitMQClient = conn
	return nil
}

func ReceiveMessageRabbitMQ() (*amqp.Delivery, error) {
	ch, err := RabbitMQClient.Channel()
	if err != nil {
		return nil, err
	}
	q, err := ch.QueueDeclare(RabbitMQQueue, false, false, true, false, nil)
	if err != nil {
		return nil, err
	}
	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	msg := <-msgs
	return &msg, nil
}
