package server

import (
	"crypto/rand"
	"encoding/base64"
	"radgifa/ent"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/microcosm-cc/bluemonday"
	"go.uber.org/zap"
)

type NewQuestionnaireRequest struct {
	Title       string `json:"title" validate:"required,min=1,max=200,no_whitespace_only"`
	Description string `json:"description" validate:"omitempty,max=1000"`
}

type NewMemberRequest struct {
	Action           string `json:"action" validate:"required,oneof=login register"`
	UniqueIdentifier string `json:"unique_identifier" validate:"required,min=3,max=32,username_format"`
	DisplayName      string `json:"display_name" validate:"omitempty,min=1,max=100"`
	Passcode         string `json:"passcode" validate:"omitempty,len=8"`
}

type NewQuestionRequest struct {
	Theme string `json:"theme" validate:"omitempty,max=255"`
	Text  string `json:"text" validate:"required,min=1"`
}

func (m *NewMemberRequest) Sanitize() {
	p := bluemonday.StrictPolicy()
	m.Action = strings.ToLower(strings.TrimSpace(m.Action))
	m.UniqueIdentifier = strings.ToLower(strings.TrimSpace(m.UniqueIdentifier))
	m.DisplayName = pgx.Identifier{strings.TrimSpace(p.Sanitize(m.DisplayName))}.Sanitize()
}

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
		log := GetLogger(c)
		log.Error("failed to create questionnaire",
			zap.String("user_id", userID.String()),
			zap.String("title", nq.Title),
			zap.Error(err))
		return c.JSON(500, map[string]string{"error": "could not create questionnaire"})
	}
	return c.JSON(201, questionnaire.ID)
}

func (s *Server) createQuestionnaireMember(c echo.Context) error {
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

	ctx := c.Request().Context()
	log := GetLogger(c)

	if memberReq.Action == "login" {
		if memberReq.Passcode == "" {
			return c.JSON(400, map[string]interface{}{
				"error": "validation failed",
				"details": map[string]string{
					"Passcode": "Passcode is required for login",
				},
			})
		}

		member, err := s.service.ValidateMemberCredentials(memberReq.UniqueIdentifier, memberReq.Passcode, ctx)
		if err != nil {
			log.Warn("member login attempt failed",
				zap.String("unique_identifier", memberReq.UniqueIdentifier),
				zap.String("questionnaire_id", questionnaireID.String()),
				zap.Error(err))
			return c.JSON(401, map[string]string{"error": "invalid credentials"})
		}

		if member.Edges.Questionnaire == nil {
			member, err = s.service.GetMemberWithQuestionnaire(member.ID, ctx)
			if err != nil {
				log.Error("failed to load member questionnaire",
					zap.String("member_id", member.ID.String()),
					zap.Error(err))
				return c.JSON(500, map[string]string{"error": "internal server error"})
			}
		}

		claims := &JWTClaims{
			EntityId:   member.ID.String(),
			EntityType: "member",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}
		jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		t, err := jwtToken.SignedString(jwtSecret)
		if err != nil {
			return c.JSON(500, map[string]string{"error": "could not generate token"})
		}

		return c.JSON(200, map[string]interface{}{
			"token":     t,
			"type":      "member",
			"member_id": member.ID,
			"iat":       claims.RegisteredClaims.IssuedAt.String(),
			"exp":       claims.RegisteredClaims.ExpiresAt.String(),
		})
	} else {
		isAvailable, err := s.service.IsMemberIdentifierAvailable(questionnaireID, memberReq.UniqueIdentifier, ctx)
		if err != nil {
			log.Error("failed to check member identifier availability",
				zap.String("unique_identifier", memberReq.UniqueIdentifier),
				zap.String("questionnaire_id", questionnaireID.String()),
				zap.Error(err))
			return c.JSON(500, map[string]string{"error": "could not validate identifier"})
		}
		if !isAvailable {
			return c.JSON(409, map[string]interface{}{
				"error": "validation failed",
				"details": map[string]string{
					"UniqueIdentifier": "Identifier is already taken in this questionnaire",
				},
			})
		}

		var member *ent.Member
		var passcode string
		if userID != uuid.Nil {
			member, err = s.service.CreateMember(userID, questionnaireID, memberReq.UniqueIdentifier, memberReq.DisplayName, ctx)
			if err != nil {
				log.Error("failed to create authenticated member",
					zap.String("user_id", userID.String()),
					zap.String("questionnaire_id", questionnaireID.String()),
					zap.String("unique_identifier", memberReq.UniqueIdentifier),
					zap.String("display_name", memberReq.DisplayName),
					zap.Error(err))
				return c.JSON(500, map[string]string{"error": "could not create member"})
			}
			return c.JSON(201, map[string]interface{}{
				"member_id": member.ID,
			})
		} else {
			member, passcode, err = s.service.CreateAnonymousMember(questionnaireID, memberReq.UniqueIdentifier, memberReq.DisplayName, ctx)
			if err != nil {
				log.Error("failed to create anonymous member",
					zap.String("questionnaire_id", questionnaireID.String()),
					zap.String("unique_identifier", memberReq.UniqueIdentifier),
					zap.String("display_name", memberReq.DisplayName),
					zap.Error(err))
				return c.JSON(500, map[string]string{"error": "could not create anonymous member"})
			}
			return c.JSON(201, map[string]interface{}{
				"member_id":         member.ID,
				"unique_identifier": member.UniqueIdentifier,
				"passcode":          passcode,
				"message":           "Save this passcode to access your member account",
			})
		}
	}
}

func generateInvitationToken() (string, error) {
	bytes := make([]byte, 16)
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

	qID, err := uuid.Parse(questionnaireID)
	if err != nil {
		return c.JSON(400, map[string]string{"error": "invalid questionnaire ID"})
	}

	ctx := c.Request().Context()
	questionnaire, err := s.service.GetQuestionnaire(qID, ctx)
	if err != nil {
		log := GetLogger(c)
		log.Error("failed to get questionnaire for invitation",
			zap.String("questionnaire_id", qID.String()),
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return c.JSON(404, map[string]string{"error": "questionnaire not found"})
	}

	u, err := questionnaire.QueryOwner().Only(ctx)
	if err != nil {
		return c.JSON(500, map[string]string{"error": "could not verify questionnaire owner"})
	}
	if u.ID.String() != userID.String() {
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

func (s *Server) checkMemberIdentifierAvailability(c echo.Context) error {
	req := new(CheckAvailabilityRequest)
	if err := BindAndValidate(c, req); err != nil {
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

	ctx := c.Request().Context()
	isAvailable, err := s.service.IsMemberIdentifierAvailable(questionnaireID, req.Value, ctx)
	if err != nil {
		log := GetLogger(c)
		log.Error("failed to check member identifier availability",
			zap.String("unique_identifier", req.Value),
			zap.String("questionnaire_id", questionnaireID.String()),
			zap.Error(err))
		return c.JSON(500, map[string]string{"error": "could not check identifier availability"})
	}

	return c.JSON(200, map[string]interface{}{
		"available":         isAvailable,
		"unique_identifier": req.Value,
	})
}

func (s *Server) createNewQuestion(c echo.Context) error {
	entityIDStr, entityType, err := GetValuesFromToken(c)
	if err != nil || entityType != "user" {
		return c.JSON(401, map[string]string{"error": "unauthorized, invalid token"})
	}
	_, err = uuid.Parse(entityIDStr)
	if err != nil {
		return c.JSON(401, map[string]string{"error": "unauthorized, invalid token"})
	}
	questionnaireID := c.Param("id")
	nq := new(NewQuestionRequest)
	if err := BindAndValidate(c, nq); err != nil {
		return err
	}
	ctx := c.Request().Context()
	questionnaireUUID, err := uuid.Parse(questionnaireID)
	if err != nil {
		return c.JSON(400, map[string]string{"error": "invalid questionnaire ID"})
	}
	q, err := s.service.GetQuestionnaire(questionnaireUUID, ctx)
	if err != nil {
		return c.JSON(404, map[string]string{"error": "questionnaire not found"})
	}
	u, err := q.QueryOwner().Only(ctx)
	if err != nil {
		return c.JSON(500, map[string]string{"error": "could not verify questionnaire owner"})
	}
	if u.ID.String() != entityIDStr {
		return c.JSON(403, map[string]string{"error": "not authorized to add questions to this questionnaire"})
	}

	question, err := s.service.CreateNewQuestion(questionnaireUUID, nq.Text, nq.Theme, ctx)
	if err != nil {
		return c.JSON(500, map[string]string{"error": "could not create question"})
	}
	return c.JSON(201, map[string]any{"question_id": question.ID})
}

// getUserQuestionnaires returns all questionnaires owned by the authenticated user
func (s *Server) getUserQuestionnaires(c echo.Context) error {
	entityIDStr, entityType, err := GetValuesFromToken(c)
	if err != nil || entityType != "user" {
		return c.JSON(401, map[string]string{"error": "unauthorized, invalid token"})
	}
	userID, err := uuid.Parse(entityIDStr)
	if err != nil {
		return c.JSON(401, map[string]string{"error": "unauthorized, invalid token"})
	}

	ctx := c.Request().Context()
	questionnaires, err := s.service.GetUserQuestionnaires(userID, ctx)
	if err != nil {
		log := GetLogger(c)
		log.Error("failed to get user questionnaires",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return c.JSON(500, map[string]string{"error": "could not get questionnaires"})
	}

	return c.JSON(200, questionnaires)
}

// getQuestionnaireDetails returns questionnaire details if user is owner or member
func (s *Server) getQuestionnaireDetails(c echo.Context) error {
	questionnaireID := c.Param("id")
	qID, err := uuid.Parse(questionnaireID)
	if err != nil {
		return c.JSON(400, map[string]string{"error": "invalid questionnaire ID"})
	}

	entityIDStr, entityType, err := GetValuesFromToken(c)
	if err != nil {
		return c.JSON(401, map[string]string{"error": "unauthorized, invalid token"})
	}

	ctx := c.Request().Context()
	questionnaire, err := s.service.GetQuestionnaireWithDetails(qID, ctx)
	if err != nil {
		return c.JSON(404, map[string]string{"error": "questionnaire not found"})
	}

	// Check permissions: only owner or members can see details
	userID, _ := uuid.Parse(entityIDStr)
	isOwner := entityType == "user" && questionnaire.Edges.Owner.ID == userID
	isMember := false

	switch entityType {
	case "user":
		// Check if user is a member
		_, err := s.service.GetMemberByUserAndQuestionnaire(userID, qID, ctx)
		isMember = err == nil
	case "member":
		// Check if this member belongs to this questionnaire
		member, err := s.service.GetMemberWithQuestionnaire(userID, ctx)
		isMember = err == nil && member.Edges.Questionnaire.ID == qID
	}

	if !isOwner && !isMember {
		return c.JSON(403, map[string]string{"error": "forbidden"})
	}

	return c.JSON(200, questionnaire)
}

// getQuestionnaireQuestions returns questions for a questionnaire if user has access
func (s *Server) getQuestionnaireQuestions(c echo.Context) error {
	questionnaireID := c.Param("id")
	qID, err := uuid.Parse(questionnaireID)
	if err != nil {
		return c.JSON(400, map[string]string{"error": "invalid questionnaire ID"})
	}

	entityIDStr, entityType, err := GetValuesFromToken(c)
	if err != nil {
		return c.JSON(401, map[string]string{"error": "unauthorized, invalid token"})
	}

	ctx := c.Request().Context()
	// First check if questionnaire exists and user has access
	questionnaire, err := s.service.GetQuestionnaireWithDetails(qID, ctx)
	if err != nil {
		return c.JSON(404, map[string]string{"error": "questionnaire not found"})
	}

	userID, _ := uuid.Parse(entityIDStr)
	isOwner := entityType == "user" && questionnaire.Edges.Owner.ID == userID
	isMember := false

	if entityType == "user" {
		_, err := s.service.GetMemberByUserAndQuestionnaire(userID, qID, ctx)
		isMember = err == nil
	} else if entityType == "member" {
		member, err := s.service.GetMemberWithQuestionnaire(userID, ctx)
		isMember = err == nil && member.Edges.Questionnaire.ID == qID
	}

	if !isOwner && !isMember {
		return c.JSON(403, map[string]string{"error": "forbidden"})
	}

	questions, err := s.service.GetQuestionnaireQuestions(qID, ctx)
	if err != nil {
		log := GetLogger(c)
		log.Error("failed to get questionnaire questions",
			zap.String("questionnaire_id", qID.String()),
			zap.Error(err))
		return c.JSON(500, map[string]string{"error": "could not get questions"})
	}

	return c.JSON(200, questions)
}

// getQuestionnaireMembers returns members for a questionnaire if user is owner
func (s *Server) getQuestionnaireMembers(c echo.Context) error {
	questionnaireID := c.Param("id")
	qID, err := uuid.Parse(questionnaireID)
	if err != nil {
		return c.JSON(400, map[string]string{"error": "invalid questionnaire ID"})
	}

	entityIDStr, entityType, err := GetValuesFromToken(c)
	if err != nil || entityType != "user" {
		return c.JSON(401, map[string]string{"error": "unauthorized, invalid token"})
	}

	userID, err := uuid.Parse(entityIDStr)
	if err != nil {
		return c.JSON(401, map[string]string{"error": "unauthorized, invalid token"})
	}

	ctx := c.Request().Context()
	questionnaire, err := s.service.GetQuestionnaireWithDetails(qID, ctx)
	if err != nil {
		return c.JSON(404, map[string]string{"error": "questionnaire not found"})
	}

	// Only owner can see members
	if questionnaire.Edges.Owner.ID != userID {
		return c.JSON(403, map[string]string{"error": "forbidden"})
	}

	members, err := s.service.GetQuestionnaireMembers(qID, ctx)
	if err != nil {
		log := GetLogger(c)
		log.Error("failed to get questionnaire members",
			zap.String("questionnaire_id", qID.String()),
			zap.Error(err))
		return c.JSON(500, map[string]string{"error": "could not get members"})
	}

	return c.JSON(200, members)
}

// getMemberAnswers returns answers by the authenticated member/user for a questionnaire
func (s *Server) getMemberAnswers(c echo.Context) error {
	questionnaireID := c.Param("id")
	qID, err := uuid.Parse(questionnaireID)
	if err != nil {
		return c.JSON(400, map[string]string{"error": "invalid questionnaire ID"})
	}

	entityIDStr, entityType, err := GetValuesFromToken(c)
	if err != nil {
		return c.JSON(401, map[string]string{"error": "unauthorized, invalid token"})
	}

	ctx := c.Request().Context()
	var memberID uuid.UUID

	switch entityType {
	case "user":
		userID, _ := uuid.Parse(entityIDStr)
		member, err := s.service.GetMemberByUserAndQuestionnaire(userID, qID, ctx)
		if err != nil {
			return c.JSON(404, map[string]string{"error": "not a member of this questionnaire"})
		}
		memberID = member.ID
	case "member":
		memberID, _ = uuid.Parse(entityIDStr)
		// Verify member belongs to this questionnaire
		member, err := s.service.GetMemberWithQuestionnaire(memberID, ctx)
		if err != nil || member.Edges.Questionnaire.ID != qID {
			return c.JSON(403, map[string]string{"error": "forbidden"})
		}
	default:
		return c.JSON(401, map[string]string{"error": "invalid token type"})
	}

	answers, err := s.service.GetMemberAnswers(memberID, qID, ctx)
	if err != nil {
		log := GetLogger(c)
		log.Error("failed to get member answers",
			zap.String("member_id", memberID.String()),
			zap.String("questionnaire_id", qID.String()),
			zap.Error(err))
		return c.JSON(500, map[string]string{"error": "could not get answers"})
	}

	return c.JSON(200, answers)
}
