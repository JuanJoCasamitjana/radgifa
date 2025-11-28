package server

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
)

type LoginCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type JWTClaims struct {
	EntityId   string `json:"entity_id"`
	EntityType string `json:"type"`
	jwt.RegisteredClaims
}

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	// Replace default logger with Zap + Lumberjack rotating file
	logger := newZapLogger()
	// Add request ID middleware first
	e.Use(middleware.RequestID())

	e.Use(zapRequestLogger(logger))
	e.Use(zapInjectLogger(logger))

	// Keep recover middleware
	e.Use(middleware.Recover())

	e.POST("/register", s.RegisterHandler)
	e.POST("/login", s.loginHandler)
	// JWT Middleware
	echojwt.JWT(jwtSecret)
	//e.Use(jwtMiddleware)

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

func (s *Server) healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, s.service.Health())
}

func (s *Server) loginHandler(c echo.Context) error {
	// Implement login logic here
	creds := new(LoginCredentials)
	if err := c.Bind(creds); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	ctx := c.Request().Context()
	user, err := s.service.ValidateUserCredentials(creds.Username, creds.Password, ctx)
	if err != nil {
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
