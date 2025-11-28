package server

import (
	"github.com/labstack/echo/v4"
)

type NewUserRequest struct {
	Name        string `json:"name" validate:"required"`
	DisplayName string `json:"display_name"`
	Username    string `json:"username" validate:"required"`
	Password    string `json:"password" validate:"required"`
}

func (s *Server) RegisterHandler(c echo.Context) error {
	nuser := new(NewUserRequest)
	if err := c.Bind(nuser); err != nil {
		return c.JSON(400, map[string]string{"error": "invalid request"})
	}
	if err := c.Validate(nuser); err != nil {
		return c.JSON(400, map[string]string{"error": "validation failed"})
	}
	_, err := s.service.CreateUser(nuser.Name, nuser.DisplayName, nuser.Username, nuser.Password, c.Request().Context())
	if err != nil {
		return c.JSON(500, map[string]string{"error": "could not register user"})
	}

	return c.JSON(201, map[string]string{"message": "user registered successfully"})
}
