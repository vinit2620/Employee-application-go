package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// Connect to PostgreSQL
	ConnectDB()

	// Register API endpoints
	http.HandleFunc("/employees", EmployeesHandler)
	http.HandleFunc("/employees/", EmployeeByIDHandler)

	// Health check
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Application is healthy!")
	})

	// Database readiness check
	http.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := DB.Ping(r.Context()); err != nil {
			http.Error(w, "Database not ready", http.StatusServiceUnavailable)
			return
		}

		fmt.Fprintln(w, "Application and database are ready!")
	})

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Employee application running on port", port)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
