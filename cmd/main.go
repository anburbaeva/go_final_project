package main

import (
	"log"
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
	defaultPort        = "7540"
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

	portValue := viper.Get("Port")
	var port string

	if portValue != nil {
		p, ok := portValue.(string)
		if !ok {
			log.Fatalf("Ошибка: значение порта не является строкой")
		}
		port = p
	} else {
		port = defaultPort
	}
	if port == defaultPort {
		logrus.Infof(serverStartMessage, port)
	} else {
		logrus.Infof(serverStartMessage+" с конфигом %+v", port, viper.AllSettings())
	}

	db, err := repository.GetDB()
	if err != nil {
		log.Fatalf("Ошибка при инициализации базы данных: %v", err)
	}
	newRepository := repository.NewRepository(db)
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
	viper.AddConfigPath("storage")
	viper.SetConfigName("storage")
	return viper.ReadInConfig()
}
