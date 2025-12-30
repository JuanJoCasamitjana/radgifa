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
	echoSwagger "github.com/swaggo/echo-swagger"
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
	Username string `json:"username" validate:"required,min=3,max=32" example:"johndoe"`
	Password string `json:"password" validate:"required,min=8" example:"password123"`
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

	e.Validator = NewValidator()

	logger := newZapLogger()

	e.Use(middleware.RequestID())

	e.Use(zapRequestLogger(logger))
	e.Use(zapInjectLogger(logger))

	// Keep recover middleware
	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://*", "http://*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	authRateLimiter := middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(reqs_sec))

	e.POST("/register", s.RegisterHandler, authRateLimiter)
	e.POST("/login", s.loginHandler, authRateLimiter)

	e.POST("/check/username", s.checkUsernameAvailability)
	e.POST("/check/member/:token", s.checkMemberIdentifierAvailability)

	e.GET("/join/:token/info", s.getQuestionnaireInfoFromToken)
	e.POST("/join/:token/info", s.getQuestionnaireInfoFromToken)
	e.POST("/join/:token", s.createQuestionnaireMember).Name = "join-questionnaire"

	api := e.Group("/api")
	jwtMiddleware := echojwt.JWT(jwtSecret)
	api.Use(jwtMiddleware)

	// Questionnaire endpoints
	api.GET("/questionnaires", s.getUserQuestionnaires)
	api.POST("/questionnaires", s.createQuestionnaire)
	api.GET("/questionnaires/:id", s.getQuestionnaireDetails)
	api.PUT("/questionnaires/:id", s.updateQuestionnaire)
	api.DELETE("/questionnaires/:id", s.deleteQuestionnaire)
	api.POST("/questionnaires/:id/publish", s.publishQuestionnaire)
	api.GET("/questionnaires/:id/questions", s.getQuestionnaireQuestions)
	api.GET("/questionnaires/:id/members", s.getQuestionnaireMembers)
	api.GET("/questionnaires/:id/my-answers", s.getMemberAnswers)
	api.POST("/questionnaires/:id/invite", s.generateQuestionnaireInvitation)
	api.POST("/questionnaires/:id/question", s.createNewQuestion)
	api.PUT("/questionnaires/:questionnaireId/questions/:questionId", s.updateQuestion)
	api.DELETE("/questionnaires/:questionnaireId/questions/:questionId", s.deleteQuestion)

	// Question endpoints
	api.POST("/question/:id", s.newQuestionAnswer)

	e.GET("/health", s.healthHandler)

	// Swagger endpoint
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Serve static frontend files
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:   "frontend/dist",
		Index:  "index.html",
		Browse: false,
		HTML5:  true,
	}))

	return e
}

func (s *Server) HelloWorldHandler(c echo.Context) error {

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

// healthHandler returns the health status of the API
// @Summary Health check
// @Description Get the health status of the API
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string "Health status"
// @Router /health [get]
func (s *Server) HealthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, s.service.Health())
}

func newZapLogger() *zap.Logger {

	fileSink := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    50, // MB
		MaxBackups: 7,
		MaxAge:     14, // days
		Compress:   true,
	})

	encCfg := zap.NewProductionEncoderConfig()
	encCfg.TimeKey = "ts"
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	jsonEncoder := zapcore.NewJSONEncoder(encCfg)

	fileCore := zapcore.NewCore(jsonEncoder, fileSink, zapcore.InfoLevel)
	consoleCore := zapcore.NewCore(jsonEncoder, zapcore.AddSync(os.Stdout), zapcore.InfoLevel)

	core := zapcore.NewTee(fileCore, consoleCore)
	return zap.New(core)
}

func zapRequestLogger(logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			start := time.Now()
			requestID := c.Response().Header().Get(echo.HeaderXRequestID)

			logger.Info("request start",
				zap.String("request_id", requestID),
				zap.String("method", req.Method),
				zap.String("path", c.Path()),
				zap.String("remote_ip", c.RealIP()),
			)

			err := next(c)
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

type ctxLoggerKey struct{}

func zapInjectLogger(base *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			requestID := c.Response().Header().Get(echo.HeaderXRequestID)
			reqLogger := base.With(
				zap.String("request_id", requestID),
				zap.String("method", req.Method),
				zap.String("path", c.Path()),
			)
			ctx := req.Context()
			ctx = context.WithValue(ctx, ctxLoggerKey{}, reqLogger)
			c.SetRequest(req.WithContext(ctx))
			return next(c)
		}
	}
}

func GetLogger(c echo.Context) *zap.Logger {
	if v := c.Request().Context().Value(ctxLoggerKey{}); v != nil {
		if zl, ok := v.(*zap.Logger); ok && zl != nil {
			return zl
		}
	}
	l, _ := zap.NewDevelopment()
	return l
}
