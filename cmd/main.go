package main

import (
	"os"

	app "github.com/anburbaeva/go_final_project"
	"github.com/anburbaeva/go_final_project/handler"
	"github.com/anburbaeva/go_final_project/repository"
	"github.com/anburbaeva/go_final_project/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	_ "modernc.org/sqlite"
)

func main() {
	checkDBDir()
	gin.SetMode(gin.ReleaseMode)

	if err := initConfig(); err != nil {
		logrus.Fatal(err)
	}

	port := app.EnvPORT("TODO_PORT")
	repo := repository.NewRepository(repository.GetDB())
	srvr := service.NewService(repo)
	handlers := handler.NewHandler(srvr)
	serv := new(app.Server)
	err := serv.Run(port, handlers.InitRoutes())
	if err != nil {
		logrus.Fatalf("ошибка в создании таблицы: %v", err)
	}
}

func checkDBDir() {
	dirName := "db"
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		err := os.Mkdir(dirName, 0700)
		if err != nil {
			logrus.Fatalf("ошибка в создании таблицы: %v", err)
			return
		}

	}
}

func initConfig() error {
	viper.AddConfigPath("config")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	return viper.ReadInConfig()
}
