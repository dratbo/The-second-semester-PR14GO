package httpapi

import (
	"encoding/json"
	"net/http"

	"example.com/pz14-rabbitmq/internal/jobs"
	"example.com/pz14-rabbitmq/internal/rabbitsetup"
	"github.com/google/uuid"
)

func (h *Handler) ProcessTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		TaskID string `json:"task_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if body.TaskID == "" {
		http.Error(w, "task_id is required", http.StatusBadRequest)
		return
	}

	if _, err := h.service.GetTaskByID(r.Context(), body.TaskID); err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	job := jobs.NewProcessTaskJob(body.TaskID, uuid.New().String())

	if err := jobs.Publish(h.ch, rabbitsetup.JobsQueue, job); err != nil {
		http.Error(w, "failed to enqueue job", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "accepted",
		"task_id": body.TaskID,
	})
}
