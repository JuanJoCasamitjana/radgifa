package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/time/rate"
)

const (
	defaultRequestsPerSecond rate.Limit = 10
)

var (
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
	reqs_sec  = setRequestsPerSecondLimit()
)

type LoginCredentials struct {
	Username string `json:"username" validate:"required,min=3,max=32"`
	Password string `json:"password" validate:"required,min=8"`
}

type JWTClaims struct {
	EntityId   string `json:"entity_id"`
	EntityType string `json:"type"`
	jwt.RegisteredClaims
}

func GetValuesFromToken(c echo.Context) (string, string, error) {
	token, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return "", "", fmt.Errorf("unauthorized, invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", fmt.Errorf("unauthorized, invalid token")
	}
	entityIDStr, ok := claims["entity_id"].(string)
	if !ok {
		return "", "", fmt.Errorf("unauthorized, invalid token")
	}
	entityType, ok := claims["type"].(string)
	if !ok {
		return "", "", fmt.Errorf("unauthorized, invalid token")
	}
	return entityIDStr, entityType, nil
}

// validateJWTToken valida un token JWT sin middleware (para rutas públicas con auth opcional)
func validateJWTToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validar que sea exactamente HS256 como se usa en loginHandler
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()

	// Configure custom validator
	e.Validator = NewValidator()

	// Replace default logger with Zap + Lumberjack rotating file
	logger := newZapLogger()
	// Add request ID middleware first
	e.Use(middleware.RequestID())

	e.Use(zapRequestLogger(logger))
	e.Use(zapInjectLogger(logger))

	// Keep recover middleware
	e.Use(middleware.Recover())

	// Rate limiter para auth endpoints: 5 requests por minuto por IP
	authRateLimiter := middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(reqs_sec))

	e.POST("/register", s.RegisterHandler, authRateLimiter)
	e.POST("/login", s.loginHandler, authRateLimiter)

	// Public routes for availability checking
	e.POST("/check/username", s.checkUsernameAvailability)
	e.POST("/check/member/:token", s.checkMemberIdentifierAvailability)

	// Public routes (no auth required)
	e.POST("/join/:token", s.createQuestionnaireMember).Name = "join-questionnaire"

	// JWT Middleware
	api := e.Group("/api")
	jwtMiddleware := echojwt.JWT(jwtSecret)
	api.Use(jwtMiddleware)

	// Protected routes
	api.POST("/questionnaires", s.createQuestionnaire)
	api.POST("/questionnaires/:id/invite", s.generateQuestionnaireInvitation)

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://*", "http://*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	e.GET("/health", s.healthHandler)

	return e
}

func (s *Server) HelloWorldHandler(c echo.Context) error {
	// Example of levelled logging inside a handler
	log := GetLogger(c)
	log.Debug("hello world handler invoked")

	resp := map[string]string{
		"message": "Hello World",
	}

	log.Info("responding hello world", zap.Int("status", http.StatusOK))
	return c.JSON(http.StatusOK, resp)
}

func setRequestsPerSecondLimit() rate.Limit {
	rateStr := os.Getenv("REQUESTS_PER_SECOND")
	if rateStr == "" {
		return defaultRequestsPerSecond
	}
	r, err := strconv.Atoi(rateStr)
	if err != nil || r <= 0 {
		return defaultRequestsPerSecond
	}
	return rate.Limit(r)
}

func (s *Server) healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, s.service.Health())
}

// newZapLogger configures a Zap logger with dual output: console and rotating file.
func newZapLogger() *zap.Logger {
	// Rotating file sink
	fileSink := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    50, // MB
		MaxBackups: 7,
		MaxAge:     14, // days
		Compress:   true,
	})

	// JSON encoder for both file and console
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.TimeKey = "ts"
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	jsonEncoder := zapcore.NewJSONEncoder(encCfg)

	fileCore := zapcore.NewCore(jsonEncoder, fileSink, zapcore.InfoLevel)
	consoleCore := zapcore.NewCore(jsonEncoder, zapcore.AddSync(os.Stdout), zapcore.InfoLevel)

	// Tee both cores
	core := zapcore.NewTee(fileCore, consoleCore)
	return zap.New(core)
}

// zapRequestLogger is a lightweight Echo middleware that logs requests with zap.
func zapRequestLogger(logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			start := time.Now()
			requestID := c.Response().Header().Get(echo.HeaderXRequestID)

			// Before
			logger.Info("request start",
				zap.String("request_id", requestID),
				zap.String("method", req.Method),
				zap.String("path", c.Path()),
				zap.String("remote_ip", c.RealIP()),
			)

			// Process
			err := next(c)

			// After
			latency := time.Since(start)
			logger.Info("request end",
				zap.String("request_id", requestID),
				zap.Int("status", res.Status),
				zap.Int64("bytes_out", res.Size),
				zap.Duration("latency", latency),
				zap.String("method", req.Method),
				zap.String("path", c.Path()),
			)

			return err
		}
	}
}

// A context key for storing zap logger in echo.Context.
type ctxLoggerKey struct{}

// zapInjectLogger attaches a request-scoped zap logger into the request context
// so handlers can retrieve and log with levels.
func zapInjectLogger(base *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			requestID := c.Response().Header().Get(echo.HeaderXRequestID)
			// enrich with request-scoped fields
			reqLogger := base.With(
				zap.String("request_id", requestID),
				zap.String("method", req.Method),
				zap.String("path", c.Path()),
			)
			// inject into context
			ctx := req.Context()
			ctx = context.WithValue(ctx, ctxLoggerKey{}, reqLogger)
			c.SetRequest(req.WithContext(ctx))
			return next(c)
		}
	}
}

// GetLogger fetches the zap logger from the echo.Context; falls back to a no-op logger if missing.
func GetLogger(c echo.Context) *zap.Logger {
	if v := c.Request().Context().Value(ctxLoggerKey{}); v != nil {
		if zl, ok := v.(*zap.Logger); ok && zl != nil {
			return zl
		}
	}
	// Fallback to a development logger to avoid nil deref
	l, _ := zap.NewDevelopment()
	return l
}
