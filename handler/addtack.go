package handler

import (
	"net/http"

	"github.com/anburbaeva/go_final_project/model"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Error struct {
	Message string `json:"error"`
}

func NewResponseError(c *gin.Context, statusCode int, message string) {
	c.AbortWithStatusJSON(statusCode, Error{Message: message})
}

func (h *Handler) login(c *gin.Context) {
	h.service.Authorization.CheckAuth(c)
}

const (
	NewND       = "Получили объект NextDate со следующими данными: date: %s, now: %s, repeat: %s"
	NewTask     = "Получили объект Task со следующими данными: date: %s, title: %s, comment: %s, repeat: %s"
	NewSearch   = "Получен запрос на задачи с поисковым запросом: %v"
	RequestTask = "Получен запрос на задачу с id: %v"
	UpdateTask  = "Получили на обновление объект task со следующими данными: id: %s, date: %s, title: %s, comment: %s, repeat: %s"
	DeletedTask = "Получен запрос на удаление задачи с id: %v"
	TaskDone    = "Получен запрос на завершение задачи с id: %v"
)

func (h *Handler) nextDate(c *gin.Context) {
	var nd model.NextDate

	err := c.ShouldBindQuery(&nd)
	if err != nil {
		logrus.Error(err)
		NewResponseError(c, http.StatusBadRequest, err.Error())
		return
	}
	logrus.Printf(NewND, nd.Date, nd.Now, nd.Repeat)

	str, err := h.service.TodoTask.NextDate(nd)
	if err != nil {
		logrus.Error(err)
		NewResponseError(c, http.StatusBadRequest, err.Error())
		return
	}
	c.Writer.WriteHeader(200)
	c.Writer.Write([]byte(str))
}
func (h *Handler) createTask(c *gin.Context) {
	var task model.Task
	err := c.ShouldBindJSON(&task)
	if err != nil {
		logrus.Error(err)
		NewResponseError(c, http.StatusBadRequest, err.Error())
		return
	}
	logrus.Printf(NewTask, task.Date, task.Title, task.Comment, task.Repeat)

	id, err := h.service.TodoTask.CreateTask(task)
	if err != nil {
		logrus.Error(err)
		NewResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(200, gin.H{"id": id})

}
func (h *Handler) getTaskById(c *gin.Context) {
	id := c.Query("id")
	logrus.Printf(RequestTask, id)
	task, err := h.service.TodoTask.GetTaskById(id)
	if err != nil {
		logrus.Error(err)
		NewResponseError(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(200, task)
}
func (h *Handler) getTasks(c *gin.Context) {
	search := c.Query("search")
	logrus.Printf(NewSearch, search)
	list, err := h.service.TodoTask.GetTasks(search)
	if err != nil {
		logrus.Error(err)
		NewResponseError(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(200, list)
}
func (h *Handler) updateTask(c *gin.Context) {
	var task model.Task

	err := c.ShouldBindJSON(&task)
	if err != nil {
		logrus.Error(err)
		NewResponseError(c, http.StatusBadRequest, err.Error())
		return
	}

	logrus.Printf(UpdateTask, task.ID, task.Date, task.Title, task.Comment, task.Repeat)

	_, err = h.service.TodoTask.GetTaskById(task.ID)
	if err != nil {
		logrus.Error(err)
		NewResponseError(c, http.StatusBadRequest, err.Error())
		return
	}

	err = h.service.TodoTask.UpdateTask(task)
	if err != nil {
		logrus.Error(err)
		NewResponseError(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(200, gin.H{})
}
func (h *Handler) deleteTask(c *gin.Context) {
	id, _ := c.GetQuery("id")
	logrus.Printf(DeletedTask, id)
	err := h.service.TodoTask.DeleteTask(id)
	if err != nil {
		logrus.Error(err)
		NewResponseError(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(200, gin.H{})
}
func (h *Handler) taskDone(c *gin.Context) {
	id, _ := c.GetQuery("id")
	logrus.Printf(TaskDone, id)
	err := h.service.TodoTask.TaskDone(id)
	if err != nil {
		logrus.Error(err)
		NewResponseError(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(200, gin.H{})
}
