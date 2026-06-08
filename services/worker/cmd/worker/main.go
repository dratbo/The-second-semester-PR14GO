package main

import (
	"log"

	"example.com/pz14-rabbitmq/internal/amqpclient"
	"example.com/pz14-rabbitmq/services/worker/internal/consumer"
)

func main() {
	conn := amqpclient.MustConnect(amqpclient.RabbitURL())
	defer conn.Close()

	ch := amqpclient.MustChannel(conn)
	defer ch.Close()

	if err := consumer.Start(ch); err != nil {
		log.Fatalf("consumer error: %v", err)
	}
}
