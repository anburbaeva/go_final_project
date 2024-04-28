package repository

import (
	"github.com/anburbaeva/go_final_project/model"
	"github.com/jmoiron/sqlx"
)

type TodoTask interface {
	NextDate(nd model.NextDate) (string, error)
	CreateTask(task model.Task) (int64, error)
	GetTasks(search string) (model.ListTasks, error)
	GetTask(id string) (model.Task, error)
	UpdateTask(task model.Task) error
	DeleteTask(id string) error
	TaskDone(id string) error
}

type Repository struct {
	TodoTask
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		TodoTask: NewTodoTaskSqlite(db),
	}
}
