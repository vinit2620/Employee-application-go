package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type Employee struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Department string    `json:"department"`
	CreatedAt  time.Time `json:"created_at"`
}

func EmployeesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetEmployees(w, r)
	case http.MethodPost:
		CreateEmployee(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func GetEmployees(w http.ResponseWriter, r *http.Request) {
	rows, err := DB.Query(r.Context(), `
		SELECT id, name, email, department, created_at
		FROM employees ORDER BY id
	`)
	if err != nil {
		http.Error(w, "Database query failed", 500)
		return
	}
	defer rows.Close()

	employees := []Employee{}

	for rows.Next() {
		var e Employee

		err := rows.Scan(
			&e.ID,
			&e.Name,
			&e.Email,
			&e.Department,
			&e.CreatedAt,
		)

		if err != nil {
			http.Error(w, "Failed to read employee", 500)
			return
		}

		employees = append(employees, e)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Failed to read employees", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(employees)
}

func CreateEmployee(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var e Employee

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&e); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	if strings.TrimSpace(e.Name) == "" ||
		!strings.Contains(e.Email, "@") ||
		strings.TrimSpace(e.Department) == "" {
		http.Error(w, "Valid name, email and department required", 400)
		return
	}

	err := DB.QueryRow(r.Context(), `
		INSERT INTO employees (name, email, department)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, department, created_at
	`,
		e.Name, e.Email, e.Department,
	).Scan(
		&e.ID, &e.Name, &e.Email, &e.Department, &e.CreatedAt,
	)

	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			http.Error(w, "Email already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to create employee", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(e)
}

func EmployeeByIDHandler(w http.ResponseWriter, r *http.Request) {
	idText := strings.TrimPrefix(r.URL.Path, "/employees/")

	id, err := strconv.ParseInt(idText, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid employee ID", 400)
		return
	}

	switch r.Method {
	case http.MethodGet:
		GetEmployee(w, r, id)
	case http.MethodDelete:
		DeleteEmployee(w, r, id)
	default:
		http.Error(w, "Method not allowed", 405)
	}
}

func GetEmployee(w http.ResponseWriter, r *http.Request, id int64) {
	var e Employee

	err := DB.QueryRow(r.Context(), `
		SELECT id, name, email, department, created_at
		FROM employees WHERE id = $1
	`, id).Scan(
		&e.ID, &e.Name, &e.Email, &e.Department, &e.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "Employee not found", 404)
		return
	}

	if err != nil {
		http.Error(w, "Database query failed", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(e)
}

func DeleteEmployee(w http.ResponseWriter, r *http.Request, id int64) {
	result, err := DB.Exec(r.Context(),
		"DELETE FROM employees WHERE id = $1", id)

	if err != nil {
		http.Error(w, "Delete failed", 500)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "Employee not found", 404)
		return
	}

	fmt.Fprintln(w, "Employee deleted successfully!")
}
