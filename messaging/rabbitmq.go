package messaging

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/rabbitmq/amqp091-go"
	"github.com/CatalinPlesu/message-service/model"
)

type RabbitMQ struct {
	Connection *amqp091.Connection
	Channel    *amqp091.Channel
}

func NewRabbitMQ(rabbitMQURL string) (*RabbitMQ, error) {
	conn, err := amqp091.Dial(rabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	return &RabbitMQ{
		Connection: conn,
		Channel:    ch,
	}, nil
}

// Publish a message to the RabbitMQ queue
func (r *RabbitMQ) PublishMessage(queueName string, message model.MessageMin) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Declare the queue if it doesn't exist
	_, err = r.Channel.QueueDeclare(
		queueName,
		true,  // Durable
		false, // Auto delete
		false, // Exclusive
		false, // No-wait
		nil,   // Arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Publish the message to the queue
	err = r.Channel.Publish(
		"",         // Default exchange
		queueName,  // Routing key (queue name)
		false,      // Mandatory
		false,      // Immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("Published message to queue %s: %s", queueName, body)
	return nil
}

// Consume messages from the RabbitMQ queue
func (r *RabbitMQ) ConsumeMessages(queueName string, handler func(model.MessageMin)) error {
	// Declare the queue if it doesn't exist
	_, err := r.Channel.QueueDeclare(
		queueName,
		true,  // Durable
		false, // Auto delete
		false, // Exclusive
		false, // No-wait
		nil,   // Arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Consume messages from the queue
	msgs, err := r.Channel.Consume(
		queueName,
		"",
		true,  // Auto-acknowledge messages
		false, // Exclusive
		false, // No-local
		false, // No-wait
		nil,   // Arguments
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	// Process messages asynchronously
	go func() {
		for d := range msgs {
			var message model.MessageMin
			if err := json.Unmarshal(d.Body, &message); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				continue
			}
			handler(message) // Call the handler with the decoded message
		}
	}()

	log.Printf("Started consuming messages from queue %s", queueName)
	return nil
}

func (r *RabbitMQ) Close() error {
	if err := r.Channel.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %w", err)
	}
	if err := r.Connection.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}
	return nil
}
