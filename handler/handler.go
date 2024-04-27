package handler

import (
	"net/http"

	"github.com/anburbaeva/go_final_project/service"
	"github.com/spf13/viper"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Engine() *gin.Engine {
	router := gin.New()

	webDir := viper.Get("WEBDir").(string)

	router.GET("/api/nextdate", h.nextDate)

	api := router.Group("/api")
	{
		api.POST("/task", h.createTask)
		api.GET("/task", h.getTask)
		api.GET("/tasks", h.getTasks)
		api.PUT("/task", h.updateTask)
		api.POST("/task/done", h.taskDone)
		api.DELETE("/task", h.deleteTask)
	}

	apiStatic := api.Group("/static")
	{
		apiStatic.StaticFS("/", http.Dir(webDir))
	}
	router.StaticFS("/", http.Dir(webDir))

	return router
}
