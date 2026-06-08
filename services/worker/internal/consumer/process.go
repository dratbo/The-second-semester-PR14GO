package consumer

import (
	"fmt"
	"time"

	"example.com/pz14-rabbitmq/internal/jobs"
)

func processTask(job jobs.TaskJob) error {
	time.Sleep(2 * time.Second)

	if job.TaskID == "t_fail" {
		return fmt.Errorf("simulated processing error")
	}

	return nil
}
