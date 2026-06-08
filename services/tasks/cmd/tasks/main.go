package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"example.com/pz14-rabbitmq/internal/amqpclient"
	"example.com/pz14-rabbitmq/internal/rabbitsetup"
	httpapi "example.com/pz14-rabbitmq/services/tasks/internal/http"
	"example.com/pz14-rabbitmq/services/tasks/internal/service"
	"example.com/pz14-rabbitmq/services/tasks/internal/task"
)

func main() {
	port := os.Getenv("TASKS_PORT")
	if port == "" {
		port = "8082"
	}

	conn := amqpclient.MustConnect(amqpclient.RabbitURL())
	defer conn.Close()

	ch := amqpclient.MustChannel(conn)
	defer ch.Close()

	if err := rabbitsetup.DeclareQueues(ch); err != nil {
		log.Fatalf("declare queues error: %v", err)
	}
	log.Printf("queues declared: %s, %s", rabbitsetup.JobsQueue, rabbitsetup.DLQQueue)

	repo := task.NewRepo()
	taskService := service.NewTaskService(repo)
	handler := httpapi.NewHandler(taskService, ch)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"service": "tasks",
		})
	})

	mux.HandleFunc("/v1/jobs/process-task", handler.ProcessTask)
	mux.HandleFunc("/v1/tasks", handler.Tasks)
	mux.HandleFunc("/v1/tasks/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetTaskByID(w, r)
		case http.MethodPatch:
			handler.PatchTask(w, r)
		case http.MethodDelete:
			handler.DeleteTask(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	addr := ":" + port
	log.Println("tasks service started on", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
