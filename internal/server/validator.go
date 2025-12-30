package server

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// Sanitizer interface for types that can sanitize their own data
type Sanitizer interface {
	Sanitize()
}

// CustomValidator wraps go-playground validator
type CustomValidator struct {
	validator *validator.Validate
}

// Validate validates structs
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Extraer errores de validación específicos
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors := make(map[string]string)
			for _, fieldError := range validationErrors {
				fieldName := fieldError.Field()
				errorMessage := getFieldErrorMessage(fieldError)
				errors[fieldName] = errorMessage
			}
			return echo.NewHTTPError(400, map[string]interface{}{
				"error":   "validation failed",
				"details": errors,
			})
		}
		return echo.NewHTTPError(400, "validation failed")
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

// getFieldErrorMessage genera un mensaje de error personalizado basado en el tipo de validación
func getFieldErrorMessage(fe validator.FieldError) string {
	fieldName := fe.Field()
	tag := fe.Tag()
	param := fe.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", fieldName)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", fieldName, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", fieldName, param)
	case "username_format":
		return fmt.Sprintf("%s can only contain letters, numbers, hyphens and underscores", fieldName)
	case "password_strength":
		return fmt.Sprintf("%s must contain at least one uppercase letter, one lowercase letter, and one number", fieldName)
	case "no_whitespace_only":
		return fmt.Sprintf("%s cannot be empty or contain only whitespace", fieldName)
	case "max_bytes":
		return fmt.Sprintf("%s exceeds maximum allowed size", fieldName)
	default:
		return fmt.Sprintf("%s is invalid", fieldName)
	}
}

// BindAndValidate is a helper that binds JSON data to a struct, sanitizes it if possible, and validates it
func BindAndValidate[T any](c echo.Context, data *T) error {
	// Bind JSON data
	if err := c.Bind(data); err != nil {
		return echo.NewHTTPError(400, map[string]string{"error": "invalid request"})
	}

	// Sanitize if the type implements Sanitizer interface
	if sanitizer, ok := any(data).(Sanitizer); ok {
		sanitizer.Sanitize()
	}

	// Validate
	if err := c.Validate(data); err != nil {
		return err // El error ya viene formateado del CustomValidator
	}

	return nil
}
