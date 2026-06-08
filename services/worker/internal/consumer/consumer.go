package consumer

import (
	"encoding/json"
	"log"
	"os"
	"strconv"

	"example.com/pz14-rabbitmq/internal/jobs"
	"example.com/pz14-rabbitmq/internal/rabbitsetup"
	"example.com/pz14-rabbitmq/services/worker/internal/store"
	amqp "github.com/rabbitmq/amqp091-go"
)

const maxAttempts = 3

func prefetchCount() int {
	if v := os.Getenv("PREFETCH"); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil && n > 0 {
			return n
		}
	}
	return 1
}

func Start(ch *amqp.Channel) error {
	if err := rabbitsetup.DeclareQueues(ch); err != nil {
		return err
	}

	prefetch := prefetchCount()
	if err := ch.Qos(prefetch, 0, false); err != nil {
		return err
	}

	msgs, err := ch.Consume(
		rabbitsetup.JobsQueue,
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

	processed := store.NewProcessedStore()

	log.Printf("worker started, queue=%s prefetch=%d", rabbitsetup.JobsQueue, prefetch)

	for d := range msgs {
		var job jobs.TaskJob
		if err := json.Unmarshal(d.Body, &job); err != nil {
			log.Printf("bad message: %v", err)
			_ = d.Nack(false, false)
			continue
		}

		if processed.Exists(job.MessageID) {
			log.Printf("skip duplicate message_id=%s", job.MessageID)
			_ = d.Ack(false)
			continue
		}

		log.Printf("processing task_id=%s attempt=%d message_id=%s", job.TaskID, job.Attempt, job.MessageID)

		if err := processTask(job); err != nil {
			log.Printf("process error task_id=%s attempt=%d: %v", job.TaskID, job.Attempt, err)

			job.Attempt++
			if job.Attempt <= maxAttempts {
				if pubErr := jobs.Publish(ch, rabbitsetup.JobsQueue, job); pubErr != nil {
					log.Printf("retry publish error: %v", pubErr)
				} else {
					log.Printf("retry scheduled task_id=%s attempt=%d", job.TaskID, job.Attempt)
				}
				_ = d.Ack(false)
				continue
			}

			if pubErr := jobs.Publish(ch, rabbitsetup.DLQQueue, job); pubErr != nil {
				log.Printf("dlq publish error: %v", pubErr)
			} else {
				log.Printf("moved to dlq task_id=%s message_id=%s", job.TaskID, job.MessageID)
			}
			_ = d.Ack(false)
			continue
		}

		processed.MarkDone(job.MessageID)
		log.Printf("done task_id=%s message_id=%s", job.TaskID, job.MessageID)
		_ = d.Ack(false)
	}

	return nil
}
