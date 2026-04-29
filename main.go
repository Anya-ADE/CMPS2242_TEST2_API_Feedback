package main

import (
	"database/sql"
	"log"
	"net/http"
)

type application struct {
	db *sql.DB
}

func main() {
	db, err := openDB("postgres://feedback:feedback@localhost:5432/feedback?sslmode=disable")
	if err != nil {
		log.Fatalf("Cannot open database: %v", err)
	}
	defer db.Close()

	app := &application{db: db}

	mux := setupRoutes(app)

	log.Println("=== FEEDBACK FORM API ===")
	log.Println("Server starting on :4000")
	log.Println()
	log.Println("=== API Endpoints ===")
	log.Println("  POST   /api/feedback          - Create new feedback")
	log.Println("  GET    /api/feedback          - Get all feedback")
	log.Println("  GET    /api/feedback/{id}     - Get single feedback")
	log.Println("  PUT    /api/feedback/{id}     - Update feedback")
	log.Println("  DELETE /api/feedback/{id}     - Delete feedback")
	log.Println("  GET    /api/feedback/names    - List only names")
	log.Println("  GET    /api/feedback/emails   - List only emails")
	log.Println("  GET    /api/feedback/subjects - List only subjects")
	log.Println("  GET    /api/feedback/messages - List only messages")
	log.Println()
	log.Println("=== UI ===")
	log.Println("  GET    http://localhost:4000   - Feedback form")
	log.Println()
	log.Println("=== Health ===")
	log.Println("  GET    /health                - Health check")

	err = http.ListenAndServe(":4000", mux)
	log.Fatal(err)
}
