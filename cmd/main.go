package main

import (
	"os"

	app "github.com/anburbaeva/go_final_project"
	"github.com/anburbaeva/go_final_project/handler"
	"github.com/anburbaeva/go_final_project/repository"
	"github.com/anburbaeva/go_final_project/service"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	_ "modernc.org/sqlite"
)

const (
	serverStartMessage = "Сервер успешно запущен на порту %s"
)

func main() {
	err := existDBDir()
	if err != nil {
		logrus.Fatal(err)
		return
	}

	if err := initConfig(); err != nil {
		logrus.Fatal(err)
		return
	}

	port := viper.Get("Port").(string)
	if port == "7540" {
		logrus.Infof(serverStartMessage, port)
	} else {
		logrus.Infof(serverStartMessage+" с конфигом %+v", port, viper.AllSettings())
	}

	newRepository := repository.NewRepository(repository.DB())
	newService := service.NewService(newRepository)
	newHandler := handler.NewHandler(newService)
	newServer := new(app.Server)
	err = newServer.Run(port, newHandler.Engine())
	if err != nil {
		logrus.Fatalf("ошибка в создании таблицы: %v", err)
		return
	}
}

func existDBDir() error {
	dirName := "./db"
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		err := os.Mkdir(dirName, 0755)
		if err != nil {
			logrus.Fatalf("ошибка в создании таблицы: %v", err)
			return err
		}
	}
	return nil
}

func initConfig() error {
	viper.AddConfigPath("config")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
