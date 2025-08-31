package messaging

import (
	"context"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	q    amqp.Queue
}

func NewPublisher(url string, queue string) (*Publisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}
	q, err := ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}
	return &Publisher{conn: conn, ch: ch, q: q}, nil
}

func (p *Publisher) Publish(msg string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return p.ch.PublishWithContext(ctx, "", p.q.Name, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(msg),
		DeliveryMode: amqp.Persistent,
	})
}

func (p *Publisher) Close() {
	if p == nil {
		return
	}
	if err := p.ch.Close(); err != nil {
		log.Printf("rabbit channel close error: %v", err)
	}
	if err := p.conn.Close(); err != nil {
		log.Printf("rabbit conn close error: %v", err)
	}
}
