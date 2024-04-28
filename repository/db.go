package repository

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	_ "modernc.org/sqlite"
)

type Config struct {
	SQLDriver string
	DBFile    string
}

func GetDB() (*sqlx.DB, error) {
	dbName, err := CheckDb()
	if err != nil {
		return nil, fmt.Errorf("failed to check database: %w", err)
	}

	sqlDriver := viper.GetString("DB.SQLDriver")
	db, err := sqlx.Connect(sqlDriver, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

func CheckDb() (string, error) {
	dbName := viper.GetString("DB.DBFile")

	appPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	dbFile := filepath.Join(filepath.Dir(appPath), dbName)
	_, err = os.Stat(dbFile)
	if err != nil {
		err = installDB(dbName)
		if err != nil {
			return "", err
		}
	}
	return dbName, nil
}

func installDB(dbName string) error {
	db, err := sql.Open("sqlite", dbName)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	createTableSQL := viper.GetString("DB.SQLCreateTables")
	_, err = db.Exec(createTableSQL)
	if err != nil {
		return err
	}
	return nil
}
