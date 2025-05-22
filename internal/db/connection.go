package db

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() *gorm.DB {
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file")
	}

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	// if dbUser == "" || dbPassword == "" || dbHost == "" || dbPort == "" || dbName == "" {
	// 	log.Fatalf("Database environment variables are not set properly. DB_USER: %q, DB_PASSWORD: %q, DB_HOST: %q, DB_PORT: %q, DB_NAME: %q", dbUser, dbPassword, dbHost, dbPort, dbName)
	// }

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true", dbUser, dbPassword, dbHost, dbPort, dbName)
	// print the DSN for debugging
	log.Printf("Connecting to database with DSN: %s", dsn)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	log.Println("Connected to database")
	DB = db
	return db
}
