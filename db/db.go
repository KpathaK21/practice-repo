package db

import (
	"fmt"
	"log"

	"github.com/KpathaK21/practice-repo/models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var DB *gorm.DB

func Init() {
	dsn := "host=localhost port=5432 user=kunjpathak dbname=kla password=password sslmode=disable"
	var err error
	DB, err = gorm.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}

	// Set the DB variable in models package
	models.DB = DB

	// Auto migrate all models
	DB.AutoMigrate(
		&models.User{},
		&models.Course{},
		&models.Material{},
		&models.Assignment{},
		&models.Submission{},
		&models.Announcement{},
		&models.Discussion{},
		&models.DiscussionReply{},
		&models.Message{},
	)

	fmt.Println("Database connected and migrated.")
}
