package jobs

const ProcessTask = "process_task"

type TaskJob struct {
	Job       string `json:"job"`
	TaskID    string `json:"task_id"`
	Attempt   int    `json:"attempt"`
	MessageID string `json:"message_id"`
}

func NewProcessTaskJob(taskID, messageID string) TaskJob {
	return TaskJob{
		Job:       ProcessTask,
		TaskID:    taskID,
		Attempt:   1,
		MessageID: messageID,
	}
}
