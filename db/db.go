package db

import(
	"fmt"
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/KpathaK21/practice-repo/models"
)

var DB *gorm.DB

func Init() {
	dsn := "host=localhost port=5432 user=kunjpathak dbname=kla password=password sslmode=disable"
	var err error
	DB, err = gorm.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	DB.AutoMigrate(&models.User{})
	fmt.Println("Database connected and migrated.")
}
