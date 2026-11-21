package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func ConnectDB() {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "employee_db")

	u := url.URL{
		Scheme: "postgres",
		Host:   net.JoinHostPort(host, port),
		Path:   dbname,
	}
	u.User = url.UserPassword(user, password)

	query := u.Query()
	query.Set("sslmode", "disable")
	u.RawQuery = query.Encode()

	config, err := pgxpool.ParseConfig(u.String())
	if err != nil {
		log.Fatal("Database configuration error:", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	DB, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatal("Database connection error:", err)
	}

	if err := DB.Ping(ctx); err != nil {
		log.Fatal("Database ping failed:", err)
	}

	fmt.Println("Successfully connected to PostgreSQL!")

	createTable := `
	CREATE TABLE IF NOT EXISTS employees (
		id BIGSERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		email VARCHAR(150) UNIQUE NOT NULL,
		department VARCHAR(100) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = DB.Exec(ctx, createTable)
	if err != nil {
		log.Fatal("Table creation failed:", err)
	}

	fmt.Println("Employee table is ready!")
}
