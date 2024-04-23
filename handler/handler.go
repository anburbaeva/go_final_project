package handler

import (
	"net/http"
	"path"

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
	urlCss := path.Join(webDir, "css")
	urlJs := path.Join(webDir, "js")
	urlIndex := path.Join(webDir, "index.html")
	urlLogin := path.Join(webDir, "login.html")
	urlFavicon := path.Join(webDir, "favicon.ico")

	router.GET("/api/nextdate", h.nextDate)

	api := router.Group("/api")
	{
		api.POST("/task", h.createTask)
		api.GET("/task", h.getTaskById)
		api.GET("/tasks", h.getTasks)
		api.PUT("/task", h.updateTask)
		api.POST("/task/done", h.taskDone)
		api.DELETE("/task", h.deleteTask)
	}

	static := router.Group("/")
	{
		static.StaticFS("./css", http.Dir(urlCss))
		static.StaticFS("./js", http.Dir(urlJs))
	}

	router.GET("/", h.indexPage)
	router.StaticFile("/index.html", urlIndex)
	router.StaticFile("/login.html", urlLogin)
	router.StaticFile("/favicon.ico", urlFavicon)

	return router
}

func (h *Handler) indexPage(c *gin.Context) {
	url := path.Join(viper.Get("WEBDir").(string), "index.html")
	http.ServeFile(c.Writer, c.Request, url)
}
