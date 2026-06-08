package service

import (
	"context"

	"example.com/pz14-rabbitmq/services/tasks/internal/task"
)

type TaskService struct {
	repo *task.Repo
}

func NewTaskService(repo *task.Repo) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) ListTasks(ctx context.Context) ([]task.Task, error) {
	return s.repo.ListAll(), nil
}

func (s *TaskService) GetTaskByID(ctx context.Context, id string) (task.Task, error) {
	return s.repo.GetByID(id)
}

func (s *TaskService) CreateTask(ctx context.Context, title string, description *string) (task.Task, error) {
	return s.repo.Create(title, description), nil
}

func (s *TaskService) UpdateTask(ctx context.Context, id string, title *string, description *string, done *bool) (task.Task, error) {
	return s.repo.Update(id, title, description, done)
}

func (s *TaskService) DeleteTask(ctx context.Context, id string) error {
	return s.repo.Delete(id)
}
