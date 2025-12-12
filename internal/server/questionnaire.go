package server

import (
	"crypto/rand"
	"encoding/base64"
	"radgifa/ent"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/microcosm-cc/bluemonday"
)

type NewQuestionnaireRequest struct {
	Title       string `json:"title" validate:"required,min=1,max=200,no_whitespace_only"`
	Description string `json:"description" validate:"omitempty,max=1000"`
}

type NewMemberRequest struct {
	DisplayName string `json:"display_name" validate:"required,min=1,max=100"`
}

// Sanitize limpia y normaliza los datos de entrada
func (m *NewMemberRequest) Sanitize() {
	p := bluemonday.StrictPolicy()
	m.DisplayName = pgx.Identifier{strings.TrimSpace(p.Sanitize(m.DisplayName))}.Sanitize()
}

// Sanitize limpia y normaliza los datos de entrada
func (q *NewQuestionnaireRequest) Sanitize() {
	p := bluemonday.StrictPolicy()
	q.Title = pgx.Identifier{strings.TrimSpace(p.Sanitize(q.Title))}.Sanitize()
	q.Description = pgx.Identifier{strings.TrimSpace(p.Sanitize(q.Description))}.Sanitize()
}

func (s *Server) createQuestionnaire(c echo.Context) error {
	entityIDStr, entityType, err := GetValuesFromToken(c)
	if err != nil || entityType != "user" {
		return c.JSON(401, map[string]string{"error": "unauthorized, invalid token"})
	}
	userID, err := uuid.Parse(entityIDStr)
	if err != nil {
		return c.JSON(401, map[string]string{"error": "unauthorized, invalid token"})
	}
	nq := new(NewQuestionnaireRequest)
	if err := BindAndValidate(c, nq); err != nil {
		return err
	}

	ctx := c.Request().Context()

	questionnaire, err := s.service.CreateQuestionnaire(userID, nq.Title, nq.Description, ctx)
	if err != nil {
		return c.JSON(500, map[string]string{"error": "could not create questionnaire"})
	}
	return c.JSON(201, questionnaire.ID)
}

func (s *Server) createQuestionnaireMember(c echo.Context) error {
	// Verificar si hay token JWT opcional (sin requerir middleware)
	var userID uuid.UUID
	if authHeader := c.Request().Header.Get("Authorization"); authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := authHeader[7:]
			if claims, err := validateJWTToken(tokenString); err == nil {
				if claims["type"] == "user" {
					if entityIDStr, ok := claims["entity_id"].(string); ok {
						userID, _ = uuid.Parse(entityIDStr)
					}
				}
			}
		}
	}

	// Validar y obtener datos del request
	memberReq := new(NewMemberRequest)
	if err := BindAndValidate(c, memberReq); err != nil {
		return err
	}

	token := c.Param("token")
	val, err := s.kvmanager.Get([]byte(token))
	if err != nil || val == nil {
		return c.JSON(400, map[string]string{"error": "invalid or expired token"})
	}
	questionnaireID, err := uuid.Parse(string(val))
	if err != nil {
		return c.JSON(400, map[string]string{"error": "invalid or expired token"})
	}

	var member *ent.Member
	if userID != uuid.Nil {
		member, err = s.service.CreateMember(userID, questionnaireID, memberReq.DisplayName, c.Request().Context())
		if err != nil {
			return c.JSON(500, map[string]string{"error": "could not create member"})
		}
	} else {
		member, err = s.service.CreateAnonymousMember(questionnaireID, memberReq.DisplayName, c.Request().Context())
		if err != nil {
			return c.JSON(500, map[string]string{"error": "could not create anonymous member"})
		}
	}
	return c.JSON(201, member.ID)
}

func generateInvitationToken() (string, error) {
	bytes := make([]byte, 16) // 128 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func (s *Server) generateQuestionnaireInvitation(c echo.Context) error {
	entityIDStr, entityType, err := GetValuesFromToken(c)
	if err != nil || entityType != "user" {
		return c.JSON(401, map[string]string{"error": "unauthorized, invalid token"})
	}

	userID, err := uuid.Parse(entityIDStr)
	if err != nil {
		return c.JSON(401, map[string]string{"error": "unauthorized, invalid token"})
	}

	questionnaireID := c.Param("id")

	// Validar que es un UUID válido
	qID, err := uuid.Parse(questionnaireID)
	if err != nil {
		return c.JSON(400, map[string]string{"error": "invalid questionnaire ID"})
	}

	// Verificar que el usuario es dueño del cuestionario
	ctx := c.Request().Context()
	questionnaire, err := s.service.GetQuestionnaire(qID, ctx)
	if err != nil {
		return c.JSON(404, map[string]string{"error": "questionnaire not found"})
	}

	if questionnaire.Edges.Owner == nil || questionnaire.Edges.Owner.ID != userID {
		return c.JSON(403, map[string]string{"error": "not authorized to invite to this questionnaire"})
	}

	token, err := generateInvitationToken()
	if err != nil {
		return c.JSON(500, map[string]string{"error": "could not generate invitation token"})
	}

	if err := s.kvmanager.InsertWithTTL([]byte(token), []byte(questionnaireID), 86400); err != nil {
		return c.JSON(500, map[string]string{"error": "could not save invitation token"})
	}

	joinURL := c.Echo().Reverse("join-questionnaire", token)

	return c.JSON(201, map[string]interface{}{
		"token":      token,
		"expires_in": "24 hours",
		"join_url":   joinURL,
	})
}
