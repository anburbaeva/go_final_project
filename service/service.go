package service

import (
	"github.com/anburbaeva/go_final_project/model"
	"github.com/anburbaeva/go_final_project/repository"
	"github.com/gin-gonic/gin"
)

type Authorization interface {
	CheckAuth(c *gin.Context)
	ParseToken(token string) (bool, error)
}
type TodoTask interface {
	NextDate(nd model.NextDate) (string, error)
	CreateTask(task model.Task) (int64, error)
	GetTasks(search string) (model.ListTasks, error)
	GetTaskById(id string) (model.Task, error)
	UpdateTask(task model.Task) error
	TaskDone(id string) error
	DeleteTask(id string) error
}

type Service struct {
	Authorization
	TodoTask
}

func NewService(repository *repository.Repository) *Service {
	return &Service{
		Authorization: NewAuthService(repository.Auth),
		TodoTask:      NewTodoTaskService(repository.TodoTask),
	}
}
