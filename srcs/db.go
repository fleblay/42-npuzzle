package main

import (
	"fmt"
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func ConnectDB(host string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(host), &gorm.Config{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to database with host :", host)
		return nil, err
	}
	return db, nil
}

func CreateModel(db *gorm.DB) *gorm.DB {
	db.AutoMigrate(&Solution{})
	return db
}
