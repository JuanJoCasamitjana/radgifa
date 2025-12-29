package server

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error" example:"error message"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string `json:"message" example:"success message"`
}

// TokenResponse represents a JWT token response
type TokenResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// QuestionnaireIDResponse represents a questionnaire ID response
type QuestionnaireIDResponse struct {
	ID string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// QuestionIDResponse represents a question ID response
type QuestionIDResponse struct {
	QuestionID string `json:"question_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}
