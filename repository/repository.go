package repository

import (
	"github.com/anburbaeva/go_final_project/model"
	"github.com/gin-gonic/gin"

	"github.com/jmoiron/sqlx"
)

type Auth interface {
	CheckAuth(c *gin.Context)
}

type TodoTask interface {
	NextDate(nd model.NextDate) (string, error)
	CreateTask(task model.Task) (int64, error)
	GetTasks(search string) (model.ListTasks, error)
	GetTaskById(id string) (model.Task, error)
	UpdateTask(task model.Task) error
	DeleteTask(id string) error
	TaskDone(id string) error
}

type Repository struct {
	Auth
	TodoTask
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Auth:     NewAuthSqlite(db),
		TodoTask: NewTodoTaskSqlite(db),
	}
}
