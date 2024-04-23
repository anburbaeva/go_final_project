package app

import (
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	GettingPortFromEnvironment = "Получаем порт из окружения..."
	UsingDefaultPort           = "Порт не задан. Будем использовать из конфига -- "
	PortSet                    = "Порт задан -- "
	StartingServer             = "Запуск сервера... Если дальше нет ошибок, то сервер успешно запущен"
)

type Server struct {
	httpserver *http.Server
}

func (s *Server) Run(port string, handler http.Handler) error {
	logrus.Println(StartingServer)
	s.httpserver = &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}
	return s.httpserver.ListenAndServe()
}

func EnvPORT(key string) string {
	logrus.Println(GettingPortFromEnvironment)
	port := os.Getenv(key)
	if len(port) == 0 {
		port = viper.Get("Port").(string)
		logrus.Warnf(UsingDefaultPort + port)
	} else {
		logrus.Println(PortSet + port)
	}
	return port
}
