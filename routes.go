package main

import (
	"net/http"
)

func setupRoutes(app *application) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		app.writeJSON(w, http.StatusOK, envelope{"status": "available"}, nil)
	})

	mux.HandleFunc("GET /", app.serveFeedbackForm)
	mux.HandleFunc("GET /UI/", app.serveStatic)

	mux.HandleFunc("POST /api/feedback", app.createFeedback)
	mux.HandleFunc("GET /api/feedback", app.listFeedbacks)
	mux.HandleFunc("GET /api/feedback/{id}", app.getFeedback)
	mux.HandleFunc("PUT /api/feedback/{id}", app.updateFeedback)
	mux.HandleFunc("DELETE /api/feedback/{id}", app.deleteFeedback)

	mux.HandleFunc("GET /api/feedback/names", app.listNames)
	mux.HandleFunc("GET /api/feedback/emails", app.listEmails)
	mux.HandleFunc("GET /api/feedback/subjects", app.listSubjects)
	mux.HandleFunc("GET /api/feedback/messages", app.listMessages)

	return mux
}
