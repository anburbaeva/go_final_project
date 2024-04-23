package handler

import (
	"net/http"
	"os"

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

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	router.POST("/api/signin", h.login)
	router.GET("/api/nextdate", h.nextDate)

	api := router.Group("/api", h.authMiddleware)
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
		static.StaticFS("./css", http.Dir(viper.Get("WEBDir").(string)+"/css"))
		static.StaticFS("./js", http.Dir(viper.Get("WEBDir").(string)+"/js"))
	}

	router.GET("/", h.indexPage)
	router.StaticFile("/index.html", "./web/index.html")
	router.StaticFile("/login.html", "./web/login.html")
	router.StaticFile("/favicon.ico", "./web/favicon.ico")

	return router
}

func (h *Handler) indexPage(c *gin.Context) {
	if os.Getenv("TODO_PASSWORD") == "" {
		deleteCookie := &http.Cookie{
			Name:     "token",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
		}
		http.SetCookie(c.Writer, deleteCookie)
	}
	http.ServeFile(c.Writer, c.Request, "./web/index.html")
}
