package server

import (
	"strings"

	"github.com/labstack/echo/v4"
)

type NewUserRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=200,no_whitespace_only"`
	DisplayName string `json:"display_name" validate:"omitempty,min=1,max=100"`
	Username    string `json:"username" validate:"required,min=3,max=32,username_format"`
	Password    string `json:"password" validate:"required,min=8,max_bytes=72,password_strength"`
}

// Sanitize limpia y normaliza los datos de entrada
func (u *NewUserRequest) Sanitize() {
	u.Name = strings.TrimSpace(u.Name)
	u.DisplayName = strings.TrimSpace(u.DisplayName)
	u.Username = strings.ToLower(strings.TrimSpace(u.Username))
	// No sanitizamos password para preservar espacios intencionales
}

func (s *Server) RegisterHandler(c echo.Context) error {
	nuser := new(NewUserRequest)
	if err := c.Bind(nuser); err != nil {
		return c.JSON(400, map[string]string{"error": "invalid request"})
	}

	// Sanitizar datos de entrada
	nuser.Sanitize()

	if err := c.Validate(nuser); err != nil {
		return c.JSON(400, map[string]string{"error": "validation failed"})
	}
	_, err := s.service.CreateUser(nuser.Name, nuser.DisplayName, nuser.Username, nuser.Password, c.Request().Context())
	if err != nil {
		return c.JSON(500, map[string]string{"error": "could not register user"})
	}

	return c.JSON(201, map[string]string{"message": "user registered successfully"})
}
