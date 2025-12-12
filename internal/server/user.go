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
	Name        string `json:"name" validate:"required,min=1,max=200,no_whitespace_only"`
	DisplayName string `json:"display_name" validate:"max=100"`
	Username    string `json:"username" validate:"required,min=3,max=32,username_format"`
	Password    string `json:"password" validate:"required,min=8,max_bytes=72,password_strength"`
}

// Sanitize limpia y normaliza los datos de entrada
func (u *NewUserRequest) Sanitize() {
	p := bluemonday.StrictPolicy()
	u.Name = pgx.Identifier{strings.TrimSpace(p.Sanitize(u.Name))}.Sanitize()
	u.DisplayName = pgx.Identifier{strings.TrimSpace(p.Sanitize(u.DisplayName))}.Sanitize()
	u.Username = strings.ToLower(strings.TrimSpace(u.Username))
	// No sanitizamos password para preservar espacios intencionales
}

func (s *Server) RegisterHandler(c echo.Context) error {
	nuser := new(NewUserRequest)
	if err := BindAndValidate(c, nuser); err != nil {
		return err
	}

	_, err := s.service.CreateUser(nuser.Name, nuser.DisplayName, nuser.Username, nuser.Password, c.Request().Context())
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

func (s *Server) loginHandler(c echo.Context) error {
	creds := new(LoginCredentials)
	if err := BindAndValidate(c, creds); err != nil {
		return err
	}

	ctx := c.Request().Context()
	user, err := s.service.ValidateUserCredentials(creds.Username, creds.Password, ctx)
	if err != nil {
		log := GetLogger(c)
		log.Warn("login attempt failed",
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

	return c.JSON(http.StatusOK, map[string]string{"token": t, "iat": claims.RegisteredClaims.IssuedAt.String(), "exp": claims.RegisteredClaims.ExpiresAt.String()})
}
