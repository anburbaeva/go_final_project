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

const (
	StartMessage            = "Начали"
	EnvPort                 = "TODO_PORT"
	dirDBfile               = "db"
	CheckDBDir              = "Проверка существования таблицы: %v..."
	ErrCreateDirectory      = "Ошибка при создании таблицы: %v"
	CreatingDirectory       = "Таблица %v не существует. Создаем..."
	SuccessDirectoryCreated = "Таблица успешно создан: %v"
	DirectoryExists         = "Таблица существует"
	InitConfig              = "Инициализация конфигурации..."
	InitConfigDone          = "Конфигурация успешно загружена"
	ErrServerStartReason    = "Ошибка при запуске сервера: %v"
)

func main() {
	logrus.Println(StartMessage)
	checkDBDir()
	gin.SetMode(gin.ReleaseMode)

	if err := initConfig(); err != nil {
		logrus.Fatal(err)
	}
	logrus.Println(InitConfigDone)

	port := app.EnvPORT(EnvPort)
	repo := repository.NewRepository(repository.GetDB())
	srvr := service.NewService(repo)
	handlers := handler.NewHandler(srvr)
	serv := new(app.Server)
	err := serv.Run(port, handlers.InitRoutes())
	if err != nil {
		logrus.Fatalf(ErrServerStartReason, err)
	}
}

func checkDBDir() {
	dirName := dirDBfile
	logrus.Printf(CheckDBDir, dirName)
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		logrus.Warnf(CreatingDirectory, dirName)
		err := os.Mkdir(dirName, 0700)
		if err != nil {
			logrus.Fatalf(ErrCreateDirectory, err)
			return
		}
		logrus.Printf(SuccessDirectoryCreated, dirName)
	} else {
		logrus.Println(DirectoryExists)
	}
}

func initConfig() error {
	viper.AddConfigPath("config")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	logrus.Println(InitConfig)
	return viper.ReadInConfig()
}
