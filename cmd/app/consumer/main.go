package consumer

import (
	"fmt"
	"transfers-api/internal/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	cfg := config.ParseFromEnv()

	url := fmt.Sprintf(
		"amqp://%s:%s@%s:%d/",
		cfg.RabbitMQConfig.Username,
		cfg.RabbitMQConfig.Password,
		"localhost",
		cfg.RabbitMQConfig.Port,
	)

	conn, err := amqp.Dial(url)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		cfg.RabbitMQConfig.Queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	msgs, err := ch.Consume(
		cfg.RabbitMQConfig.Queue,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("Listening notifications")

	for msg := range msgs {
		fmt.Println("Notification received:", string(msg.Body))
	}
}
