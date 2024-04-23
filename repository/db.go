package repository

import (
	"database/sql"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	_ "modernc.org/sqlite"
)

type Config struct {
	SQLDriver string
	DBFile    string
}

func GetDB() *sqlx.DB {
	dbName, err := CheckDb()
	if err != nil {
		logrus.Fatal(err)
	}
	sqlDriver := viper.GetString("DB.SQLDriver")
	return sqlx.MustConnect(sqlDriver, dbName)
}

func CheckDb() (string, error) {
	dbName := EnvDBFILE("TODO_DBFILE")

	appPath, err := os.Executable()
	if err != nil {
		logrus.Fatal(err)
	}
	dbFile := filepath.Join(filepath.Dir(appPath), dbName)
	_, err = os.Stat(dbFile)
	if err != nil {
		installDB(dbName)
	}
	return dbName, nil
}

func EnvDBFILE(key string) string {
	dbName := os.Getenv(key)
	if len(dbName) == 0 {
		dbName = viper.Get("DB.DBFile").(string)
	}
	return dbName
}

func installDB(dbName string) {
	db, err := sql.Open("sqlite", dbName)
	if err != nil {
		logrus.Fatal("бд не открылось: ", err)
	}
	defer db.Close()

	createTableSQL := viper.Get("DB.SQLCreateTables").(string)
	_, err = db.Exec(createTableSQL)

	createIndexSQL := viper.Get("DB.SQLCreateIndexes").(string)
	_, err = db.Exec(createIndexSQL)

}
