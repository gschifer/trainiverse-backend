package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var DB *sql.DB

func InitDatabase() {
	envErr := godotenv.Load("../internal/config/.env")
	if envErr != nil {
		log.Fatal("Error loading .env file")
	}

	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", user, password, dbname, sslmode)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal("Failed to ping the database:", err)
	}

	log.Println("Connected to the database successfully")

	RunMigrations(DB)
}

func RunMigrations(db *sql.DB) {

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal("Error creating migration driver:", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://../internal/db/migrations",
		"postgres", driver)

	if err != nil {
		log.Fatal("Migration init failed:", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("Migration up failed:", err)
	}

	log.Println("Migrations applied successfully")
}
