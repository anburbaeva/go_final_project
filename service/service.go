package service

import (
	"github.com/anburbaeva/go_final_project/model"
	"github.com/anburbaeva/go_final_project/repository"
)

type TodoTask interface {
	NextDate(nd model.NextDate) (string, error)
	CreateTask(task model.Task) (int64, error)
	GetTasks(search string) (model.ListTasks, error)
	GetTask(id string) (model.Task, error)
	UpdateTask(task model.Task) error
	TaskDone(id string) error
	DeleteTask(id string) error
}

type Service struct {
	TodoTask
}

func NewService(repository *repository.Repository) *Service {
	return &Service{
		TodoTask: NewTodoTaskService(repository.TodoTask),
	}
}
