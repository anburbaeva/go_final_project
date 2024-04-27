package repository

import (
	"database/sql"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	_ "modernc.org/sqlite"
)

type Config struct {
	SQLDriver string
	DBFile    string
}

func DB() *sqlx.DB {
	dbName, err := CheckDb()
	if err != nil {
		logrus.Fatal(err)
		return nil
	}
	sqlDriver := viper.GetString("DB.SQLDriver")
	return sqlx.MustConnect(sqlDriver, dbName)
}

func CheckDb() (string, error) {
	dbName := viper.Get("DB.DBFile").(string)

	_, err := os.Stat(dbName)
	if err != nil {
		if os.IsNotExist(err) {
			installDB(dbName)
		} else {
			logrus.Fatal(err)
			return "", err
		}
	}
	return dbName, nil
}

func installDB(dbName string) {
	db, err := sql.Open("sqlite", dbName)
	if err != nil {
		logrus.Fatal("бд не открылось: ", err)
		return
	}
	defer func() { _ = db.Close() }()

	createTableSQL := viper.Get("DB.SQLCreateTables").(string)
	_, err = db.Exec(createTableSQL)
	if err != nil {
		logrus.Fatal("создание таблицы: ", err)
		return
	}
}
