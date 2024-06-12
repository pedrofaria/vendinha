package db

import (
	"vendinha/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func Connect() error {
	var err error
	db, err = gorm.Open(sqlite.Open("vendinha.db"), &gorm.Config{})
	if err != nil {
		return err
	}

	// Migrate the schema
	if err := db.AutoMigrate(&model.Product{}); err != nil {
		return err
	}

	return nil
}

func GetDB() *gorm.DB {
	return db
}
