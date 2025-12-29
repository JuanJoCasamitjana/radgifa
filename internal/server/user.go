package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/microcosm-cc/bluemonday"
	"go.uber.org/zap"
)

type NewUserRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=200,no_whitespace_only" example:"John Doe"`
	DisplayName string `json:"display_name" validate:"omitempty,min=1,max=100" example:"Johnny"`
	Username    string `json:"username" validate:"required,min=3,max=32,username_format" example:"johndoe"`
	Password    string `json:"password" validate:"required,min=8,max_bytes=72,password_strength" example:"password123"`
}

type CheckAvailabilityRequest struct {
	Value string `json:"value" validate:"required,min=3,max=32,username_format" example:"johndoe"`
}

func (u *NewUserRequest) Sanitize() {
	p := bluemonday.StrictPolicy()
	u.Name = pgx.Identifier{strings.TrimSpace(p.Sanitize(u.Name))}.Sanitize()
	u.DisplayName = pgx.Identifier{strings.TrimSpace(p.Sanitize(u.DisplayName))}.Sanitize()
	u.Username = strings.ToLower(strings.TrimSpace(u.Username))
	// No sanitizamos password para preservar espacios intencionales
}

// RegisterHandler registers a new user
// @Summary Register a new user
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param user body NewUserRequest true "User registration data"
// @Success 201 {object} map[string]string "User created successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 409 {object} map[string]string "Username already exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /register [post]
func (s *Server) RegisterHandler(c echo.Context) error {
	nuser := new(NewUserRequest)
	if err := BindAndValidate(c, nuser); err != nil {
		return err
	}

	// Verificar que el username esté disponible
	ctx := c.Request().Context()
	isAvailable, err := s.service.IsUsernameAvailable(nuser.Username, ctx)
	if err != nil {
		log := GetLogger(c)
		log.Error("failed to check username availability during registration",
			zap.String("username", nuser.Username),
			zap.Error(err))
		return c.JSON(500, map[string]string{"error": "could not validate username"})
	}
	if !isAvailable {
		return c.JSON(409, map[string]interface{}{
			"error": "validation failed",
			"details": map[string]string{
				"Username": "Username is already taken",
			},
		})
	}

	_, err = s.service.CreateUser(nuser.Name, nuser.DisplayName, nuser.Username, nuser.Password, ctx)
	if err != nil {
		log := GetLogger(c)
		log.Error("failed to create user",
			zap.String("username", nuser.Username),
			zap.String("name", nuser.Name),
			zap.Error(err))
		return c.JSON(500, map[string]string{"error": "could not register user"})
	}

	return c.JSON(201, map[string]string{"message": "user registered successfully"})
}

// loginHandler authenticates a user
// @Summary Login user
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body LoginCredentials true "Login credentials"
// @Success 200 {object} map[string]string "JWT token"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Invalid credentials"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /login [post]
func (s *Server) loginHandler(c echo.Context) error {
	creds := new(LoginCredentials)
	if err := BindAndValidate(c, creds); err != nil {
		return err
	}

	ctx := c.Request().Context()
	log := GetLogger(c)

	user, err := s.service.ValidateUserCredentials(creds.Username, creds.Password, ctx)
	if err != nil {
		log.Warn("user login attempt failed",
			zap.String("username", creds.Username),
			zap.Error(err))
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
	}

	claims := &JWTClaims{
		EntityId:   user.ID.String(),
		EntityType: "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString(jwtSecret)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not generate token"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"token": t,
		"type":  "user",
		"iat":   claims.RegisteredClaims.IssuedAt.String(),
		"exp":   claims.RegisteredClaims.ExpiresAt.String(),
	})
}

// checkUsernameAvailability checks if a username is available
// @Summary Check username availability
// @Description Check if a username is available for registration
// @Tags auth
// @Accept json
// @Produce json
// @Param availability body CheckAvailabilityRequest true "Username to check"
// @Success 200 {object} map[string]interface{} "Username availability status"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /check/username [post]
func (s *Server) checkUsernameAvailability(c echo.Context) error {
	req := new(CheckAvailabilityRequest)
	if err := BindAndValidate(c, req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	isAvailable, err := s.service.IsUsernameAvailable(req.Value, ctx)
	if err != nil {
		log := GetLogger(c)
		log.Error("failed to check username availability",
			zap.String("username", req.Value),
			zap.Error(err))
		return c.JSON(500, map[string]string{"error": "could not check username availability"})
	}

	return c.JSON(200, map[string]interface{}{
		"available": isAvailable,
		"username":  req.Value,
	})
}
