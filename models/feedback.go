package models

type Feedback struct {
	ID      int    `json:"id"`
	Name    int    `json:"name"`
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}
