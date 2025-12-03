package server

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// CustomValidator wraps go-playground validator
type CustomValidator struct {
	validator *validator.Validate
}

// Validate validates structs
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(400, err.Error())
	}
	return nil
}

// NewValidator creates a new validator with custom validation rules
func NewValidator() *CustomValidator {
	v := validator.New()

	// Register custom validations
	v.RegisterValidation("password_strength", validatePasswordStrength)
	v.RegisterValidation("username_format", validateUsernameFormat)
	v.RegisterValidation("no_whitespace_only", validateNoWhitespaceOnly)
	v.RegisterValidation("max_bytes", validateMaxBytes)

	return &CustomValidator{validator: v}
}

// validatePasswordStrength verifica que la contraseña tenga:
// - Al menos 1 mayúscula
// - Al menos 1 minúscula
// - Al menos 1 número
func validatePasswordStrength(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	var (
		hasUpper  bool
		hasLower  bool
		hasNumber bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		}
	}

	return hasUpper && hasLower && hasNumber
}

// validateUsernameFormat verifica que el username sea alfanumérico con guiones/guiones bajos
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func validateUsernameFormat(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	return usernameRegex.MatchString(username)
}

// validateNoWhitespaceOnly verifica que el campo no sea solo espacios
func validateNoWhitespaceOnly(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return strings.TrimSpace(value) != ""
}

// validateMaxBytes verifica que el campo no exceda el límite de bytes
func validateMaxBytes(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	maxBytes := fl.Param()

	// Parse the parameter
	var limit int
	if maxBytes == "72" {
		limit = 72
	} else {
		return true // Si no se especifica, no validamos
	}

	return len([]byte(value)) <= limit
}
