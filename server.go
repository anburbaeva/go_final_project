package app

import (
	"net/http"
	"os"

	"github.com/spf13/viper"
)

type Server struct {
	httpserver *http.Server
}

func (s *Server) Run(port string, handler http.Handler) error {
	s.httpserver = &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}
	return s.httpserver.ListenAndServe()
}

func EnvPORT(key string) string {
	port := os.Getenv(key)
	if len(port) == 0 {
		port = viper.Get("Port").(string)
	}
	return port
}
