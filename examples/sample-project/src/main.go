package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type Server struct {
	db     *sql.DB
	router *http.ServeMux
}

type User struct {
	ID    int
	Email string
	Name  string
}

func NewServer(dbPath string) (*Server, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	server := &Server{
		db:     db,
		router: http.NewServeMux(),
	}

	server.setupRoutes()
	return server, nil
}

func (s *Server) setupRoutes() {
	s.router.HandleFunc("/users", s.handleGetUsers)
	s.router.HandleFunc("/users/create", s.handleCreateUser)
	s.router.HandleFunc("/health", s.handleHealth)
}

func (s *Server) handleGetUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query("SELECT id, email, name FROM users")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Email, &user.Name); err != nil {
			log.Println(err)
			continue
		}
		users = append(users, user)
	}

	fmt.Fprintf(w, "Found %d users\n", len(users))
}

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	email := r.FormValue("email")
	name := r.FormValue("name")

	_, err := s.db.Exec("INSERT INTO users (email, name) VALUES (?, ?)", email, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "User created successfully\n")
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK\n")
}

func (s *Server) Start(addr string) error {
	log.Printf("Server starting on %s\n", addr)
	return http.ListenAndServe(addr, s.router)
}

func main() {
	server, err := NewServer(":memory:")
	if err != nil {
		log.Fatal(err)
	}

	if err := server.Start(":8080"); err != nil {
		log.Fatal(err)
	}
}
