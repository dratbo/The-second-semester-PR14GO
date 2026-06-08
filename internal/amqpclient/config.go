package amqpclient

import "os"

const (
	defaultRabbitURL = "amqp://guest:guest@localhost:5672/"
)

func RabbitURL() string {
	if url := os.Getenv("RABBIT_URL"); url != "" {
		return url
	}
	return defaultRabbitURL
}
