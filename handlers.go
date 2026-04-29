package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func (app *application) serveFeedbackForm(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, "UI/index.html")
}

func (app *application) serveStatic(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	http.ServeFile(w, r, path)
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data interface{}, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	err := json.NewDecoder(r.Body).Decode(dst)
	if err != nil {
		return err
	}
	return nil
}

func (app *application) badRequest(w http.ResponseWriter, message string) {
	app.writeJSON(w, http.StatusBadRequest, map[string]string{"error": message}, nil)
}

func (app *application) notFound(w http.ResponseWriter) {
	app.writeJSON(w, http.StatusNotFound, map[string]string{"error": "resource not found"}, nil)
}

func (app *application) serverError(w http.ResponseWriter, err error) {
	app.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}, nil)
}

func parseID(r *http.Request, param string) (int64, error) {
	idStr := r.PathValue(param)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func validateName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("name is required")
	}
	if len(name) < 2 {
		return fmt.Errorf("name must be at least 2 characters")
	}
	if len(name) > 100 {
		return fmt.Errorf("name must be less than 100 characters")
	}
	matched, _ := regexp.MatchString("^[a-zA-Z\\s\\-']+$", name)
	if !matched {
		return fmt.Errorf("name can only contain letters, spaces, hyphens, and apostrophes")
	}
	return nil
}

func validateEmail(email string) error {
	if strings.TrimSpace(email) == "" {
		return fmt.Errorf("email is required")
	}
	if len(email) > 255 {
		return fmt.Errorf("email must be less than 255 characters")
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

func validateSubject(subject string) error {
	if strings.TrimSpace(subject) == "" {
		return fmt.Errorf("subject is required")
	}
	if len(subject) < 3 {
		return fmt.Errorf("subject must be at least 3 characters")
	}
	if len(subject) > 200 {
		return fmt.Errorf("subject must be less than 200 characters")
	}
	return nil
}

func validateMessage(message string) error {
	if strings.TrimSpace(message) == "" {
		return fmt.Errorf("message is required")
	}
	if len(message) < 5 {
		return fmt.Errorf("message must be at least 5 characters")
	}
	if len(message) > 1000 {
		return fmt.Errorf("message must be less than 1000 characters")
	}
	return nil
}

func (app *application) createFeedback(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequest(w, "Invalid request body")
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.TrimSpace(input.Email)
	input.Subject = strings.TrimSpace(input.Subject)
	input.Message = strings.TrimSpace(input.Message)

	if err := validateName(input.Name); err != nil {
		app.badRequest(w, err.Error())
		return
	}

	if err := validateEmail(input.Email); err != nil {
		app.badRequest(w, err.Error())
		return
	}

	if err := validateSubject(input.Subject); err != nil {
		app.badRequest(w, err.Error())
		return
	}

	if err := validateMessage(input.Message); err != nil {
		app.badRequest(w, err.Error())
		return
	}

	query := `INSERT INTO feedback (name, email, subject, message) VALUES ($1, $2, $3, $4) RETURNING id`
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var id int
	err = app.db.QueryRowContext(ctx, query, input.Name, input.Email, input.Subject, input.Message).Scan(&id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	feedback := map[string]interface{}{
		"id":      id,
		"name":    input.Name,
		"email":   input.Email,
		"subject": input.Subject,
		"message": input.Message,
	}

	app.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Feedback submitted successfully",
		"data":    feedback,
	}, nil)
}

func (app *application) listFeedbacks(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, name, email, subject, message FROM feedback ORDER BY id DESC`
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	rows, err := app.db.QueryContext(ctx, query)
	if err != nil {
		app.serverError(w, err)
		return
	}
	defer rows.Close()

	var feedbacks []map[string]interface{}
	for rows.Next() {
		var id int
		var name, email, subject, message string
		err := rows.Scan(&id, &name, &email, &subject, &message)
		if err != nil {
			app.serverError(w, err)
			return
		}
		feedbacks = append(feedbacks, map[string]interface{}{
			"id":      id,
			"name":    name,
			"email":   email,
			"subject": subject,
			"message": message,
		})
	}

	app.writeJSON(w, http.StatusOK, feedbacks, nil)
}

func (app *application) getFeedback(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	query := `SELECT id, name, email, subject, message FROM feedback WHERE id = $1`
	var idVal int
	var name, email, subject, message string
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	err = app.db.QueryRowContext(ctx, query, id).Scan(&idVal, &name, &email, &subject, &message)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":      idVal,
		"name":    name,
		"email":   email,
		"subject": subject,
		"message": message,
	}, nil)
}

func (app *application) updateFeedback(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	var input struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequest(w, "Invalid request body")
		return
	}

	if input.Name != "" {
		input.Name = strings.TrimSpace(input.Name)
		if err := validateName(input.Name); err != nil {
			app.badRequest(w, err.Error())
			return
		}
	}

	if input.Email != "" {
		input.Email = strings.TrimSpace(input.Email)
		if err := validateEmail(input.Email); err != nil {
			app.badRequest(w, err.Error())
			return
		}
	}

	if input.Subject != "" {
		input.Subject = strings.TrimSpace(input.Subject)
		if err := validateSubject(input.Subject); err != nil {
			app.badRequest(w, err.Error())
			return
		}
	}

	if input.Message != "" {
		input.Message = strings.TrimSpace(input.Message)
		if err := validateMessage(input.Message); err != nil {
			app.badRequest(w, err.Error())
			return
		}
	}

	query := `UPDATE feedback SET 
		name = COALESCE(NULLIF($1, ''), name), 
		email = COALESCE(NULLIF($2, ''), email), 
		subject = COALESCE(NULLIF($3, ''), subject), 
		message = COALESCE(NULLIF($4, ''), message) 
	WHERE id = $5`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	result, err := app.db.ExecContext(ctx, query, input.Name, input.Email, input.Subject, input.Message, id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		app.serverError(w, err)
		return
	}

	if rowsAffected == 0 {
		app.notFound(w)
		return
	}

	app.writeJSON(w, http.StatusOK, map[string]string{"message": "Feedback updated successfully"}, nil)
}

func (app *application) deleteFeedback(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	query := `DELETE FROM feedback WHERE id = $1`
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	result, err := app.db.ExecContext(ctx, query, id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		app.serverError(w, err)
		return
	}

	if rowsAffected == 0 {
		app.notFound(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) listNames(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, name FROM feedback ORDER BY id`
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	rows, err := app.db.QueryContext(ctx, query)
	if err != nil {
		app.serverError(w, err)
		return
	}
	defer rows.Close()

	var names []map[string]interface{}
	for rows.Next() {
		var id int
		var name string
		rows.Scan(&id, &name)
		names = append(names, map[string]interface{}{"id": id, "name": name})
	}

	app.writeJSON(w, http.StatusOK, map[string]interface{}{"names": names}, nil)
}

func (app *application) listEmails(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, email FROM feedback ORDER BY id`
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	rows, err := app.db.QueryContext(ctx, query)
	if err != nil {
		app.serverError(w, err)
		return
	}
	defer rows.Close()

	var emails []map[string]interface{}
	for rows.Next() {
		var id int
		var email string
		rows.Scan(&id, &email)
		emails = append(emails, map[string]interface{}{"id": id, "email": email})
	}

	app.writeJSON(w, http.StatusOK, map[string]interface{}{"emails": emails}, nil)
}

func (app *application) listSubjects(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, subject FROM feedback ORDER BY id`
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	rows, err := app.db.QueryContext(ctx, query)
	if err != nil {
		app.serverError(w, err)
		return
	}
	defer rows.Close()

	var subjects []map[string]interface{}
	for rows.Next() {
		var id int
		var subject string
		rows.Scan(&id, &subject)
		subjects = append(subjects, map[string]interface{}{"id": id, "subject": subject})
	}

	app.writeJSON(w, http.StatusOK, map[string]interface{}{"subjects": subjects}, nil)
}

func (app *application) listMessages(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, message FROM feedback ORDER BY id`
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	rows, err := app.db.QueryContext(ctx, query)
	if err != nil {
		app.serverError(w, err)
		return
	}
	defer rows.Close()

	var messages []map[string]interface{}
	for rows.Next() {
		var id int
		var message string
		rows.Scan(&id, &message)
		messages = append(messages, map[string]interface{}{"id": id, "message": message})
	}

	app.writeJSON(w, http.StatusOK, map[string]interface{}{"messages": messages}, nil)
}
