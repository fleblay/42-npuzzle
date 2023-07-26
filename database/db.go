package database

import (
	"errors"
	"fmt"

	"github.com/fleblay/42-npuzzle/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func ConnectDB(host string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(host), &gorm.Config{})
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to connect to database with host [%s] : %s\n", host, err.Error()))
	}
	return db, nil
}

func CreateModel(db *gorm.DB) error {
	err := db.AutoMigrate(&models.Solution{})
	return err
}
