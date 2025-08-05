package database

import (
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	// Register PostgreSQL driver via side-effects
	_ "github.com/lib/pq"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	// Register file source driver for golang-migrate via side-effects
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type DBInterface interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
}

var DB *sql.DB

func InitDatabase() {
	envErr := godotenv.Load("internal/config/.env")
	if envErr != nil {
		log.Fatal("Error loading .env file")
	}
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")
	port := os.Getenv("DB_PORT")

	var dbHost string
	if os.Getenv("RUNNING_IN_DOCKER") == "true" {
		dbHost = os.Getenv("DB_HOST_FOR_DOCKER")
	} else {
		dbHost = os.Getenv("DB_HOST")
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, port, user, password, dbname, sslmode)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	for i := 0; i < 10; i++ {
		err = DB.Ping()
		if err == nil {
			break
		}
		log.Println("Waiting for database to be ready...")
	}
	if err != nil {
		log.Fatal("Failed to connect to database after multiple retries:", err)
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
		"file://internal/db/migrations",
		"postgres", driver)

	if err != nil {
		log.Fatal("Migration init failed:", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("Migration up failed:", err)
	}

	log.Println("Migrations applied successfully")
}
