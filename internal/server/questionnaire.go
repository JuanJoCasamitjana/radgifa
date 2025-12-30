package server

import (
	"crypto/rand"
	"encoding/base64"
	"radgifa/ent"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/microcosm-cc/bluemonday"
	"go.uber.org/zap"
)

type NewQuestionnaireRequest struct {
	Title       string `json:"title" validate:"required,min=1,max=200,no_whitespace_only" example:"Best Pizza Topping"`
	Description string `json:"description" validate:"omitempty,max=1000" example:"Let's decide which pizza topping to order for the team lunch"`
}

type NewMemberRequest struct {
	Action           string `json:"action" validate:"required,oneof=login register" example:"register"`
	UniqueIdentifier string `json:"unique_identifier" validate:"required,min=3,max=32,username_format" example:"participant123"`
	DisplayName      string `json:"display_name" validate:"omitempty,min=1,max=100" example:"Anonymous Participant"`
	Passcode         string `json:"passcode" validate:"omitempty,len=8" example:"ABC12345"`
}

type NewQuestionRequest struct {
	Theme string `json:"theme" validate:"omitempty,max=255" example:"Food Preferences"`
	Text  string `json:"text" validate:"required,min=1" example:"Do you like pepperoni pizza?"`
}

type UpdateQuestionnaireRequest struct {
	Title       string `json:"title" validate:"required,min=1,max=200,no_whitespace_only" example:"Updated Pizza Topping"`
	Description string `json:"description" validate:"omitempty,max=1000" example:"Updated description for the questionnaire"`
}

type UpdateQuestionRequest struct {
	Theme string `json:"theme" validate:"omitempty,max=255" example:"Updated Food Preferences"`
	Text  string `json:"text" validate:"required,min=1" example:"Do you still like pepperoni pizza?"`
}

func (m *NewMemberRequest) Sanitize() {
	p := bluemonday.StrictPolicy()
	m.Action = strings.ToLower(strings.TrimSpace(m.Action))
	m.UniqueIdentifier = strings.ToLower(strings.TrimSpace(m.UniqueIdentifier))
	m.DisplayName = strings.TrimSpace(p.Sanitize(m.DisplayName))
}

func (q *NewQuestionnaireRequest) Sanitize() {
	p := bluemonday.StrictPolicy()
	q.Title = strings.TrimSpace(p.Sanitize(q.Title))
	q.Description = strings.TrimSpace(p.Sanitize(q.Description))
}

func (nq *NewQuestionRequest) Sanitize() {
	p := bluemonday.StrictPolicy()
	nq.Text = strings.TrimSpace(p.Sanitize(nq.Text))
	nq.Theme = strings.TrimSpace(p.Sanitize(nq.Theme))
}

func (uq *UpdateQuestionnaireRequest) Sanitize() {
	p := bluemonday.StrictPolicy()
	uq.Title = strings.TrimSpace(p.Sanitize(uq.Title))
	uq.Description = strings.TrimSpace(p.Sanitize(uq.Description))
}

func (uq *UpdateQuestionRequest) Sanitize() {
	p := bluemonday.StrictPolicy()
	uq.Text = strings.TrimSpace(p.Sanitize(uq.Text))
	uq.Theme = strings.TrimSpace(p.Sanitize(uq.Theme))
}

// createQuestionnaire creates a new questionnaire
// @Summary Create questionnaire
// @Description Create a new questionnaire for group decision making
// @Tags questionnaires
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param questionnaire body NewQuestionnaireRequest true "Questionnaire data"
// @Success 201 {object} map[string]interface{} "Questionnaire ID"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/questionnaires [post]
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

	nq.Sanitize()

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

// createQuestionnaireMember joins a user to a questionnaire using an invitation token
// @Summary Join questionnaire
// @Description Join a questionnaire using an invitation token
// @Tags questionnaires
// @Accept json
// @Produce json
// @Param token path string true "Invitation token"
// @Param member body NewMemberRequest true "Member data"
// @Success 201 {object} map[string]interface{} "Member created successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Invalid token"
// @Failure 409 {object} map[string]string "Identifier already taken"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /join/{token} [post]
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

	if userID != uuid.Nil {
		existingMember, err := s.service.GetMemberByUserAndQuestionnaire(userID, questionnaireID, ctx)
		if err == nil && existingMember != nil {
			log.Info("authenticated user already member, returning existing member",
				zap.String("user_id", userID.String()),
				zap.String("member_id", existingMember.ID.String()),
				zap.String("questionnaire_id", questionnaireID.String()))

			return c.JSON(200, map[string]interface{}{
				"member_id":      existingMember.ID,
				"already_member": true,
				"message":        "You are already a member of this questionnaire",
			})
		}
	}

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

			// Generate JWT for authenticated member
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

			return c.JSON(201, map[string]interface{}{
				"token":     t,
				"type":      "member",
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

			// Generate JWT for anonymous member
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

			return c.JSON(201, map[string]interface{}{
				"token":             t,
				"type":              "member",
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

// generateQuestionnaireInvitation generates an invitation token for a questionnaire
// @Summary Generate questionnaire invitation
// @Description Generate an invitation token to allow others to join the questionnaire
// @Tags questionnaires
// @Produce json
// @Security BearerAuth
// @Param id path string true "Questionnaire ID"
// @Success 200 {object} map[string]interface{} "Invitation token and URL"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - only owner can generate invitations"
// @Failure 404 {object} map[string]string "Questionnaire not found"
// @Failure 500 {object} map[string]string "Internal server error"
// generateQuestionnaireInvitation generates an invitation token for a questionnaire
// @Summary Generate questionnaire invitation
// @Description Generate an invitation token to allow others to join the questionnaire
// @Tags questionnaires
// @Produce json
// @Security BearerAuth
// @Param id path string true "Questionnaire ID"
// @Success 200 {object} map[string]interface{} "Invitation token and URL"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - only owner can generate invitations"
// @Failure 404 {object} map[string]string "Questionnaire not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/questionnaires/{id}/invite [post]
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

// checkMemberIdentifierAvailability checks if a member identifier is available in a questionnaire
// @Summary Check member identifier availability
// @Description Check if a member identifier is available in a specific questionnaire
// @Tags questionnaires
// @Accept json
// @Produce json
// @Param token path string true "Questionnaire invitation token"
// @Param availability body CheckAvailabilityRequest true "Identifier to check"
// @Success 200 {object} map[string]interface{} "Identifier availability status"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Invalid token"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /check/member/{token} [post]
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

// getQuestionnaireInfoFromToken gets questionnaire information from an invitation token
// @Summary Get questionnaire info from token
// @Description Get basic questionnaire information using an invitation token
// @Tags questionnaires
// @Produce json
// @Param token path string true "Questionnaire invitation token"
// @Success 200 {object} map[string]interface{} "Questionnaire information"
// @Failure 400 {object} map[string]string "Invalid or expired token"
// @Failure 404 {object} map[string]string "Questionnaire not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /join/{token}/info [get]
func (s *Server) getQuestionnaireInfoFromToken(c echo.Context) error {
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
	questionnaire, err := s.service.GetQuestionnaire(questionnaireID, ctx)
	if err != nil {
		return c.JSON(404, map[string]string{"error": "questionnaire not found"})
	}

	return c.JSON(200, map[string]interface{}{
		"questionnaire_id": questionnaire.ID,
		"title":            questionnaire.Title,
		"description":      questionnaire.Description,
		"is_published":     questionnaire.IsPublished,
	})
}

// createNewQuestion creates a new question in a questionnaire
// @Summary Create new question
// @Description Create a new question in a specific questionnaire (only owner can do this)
// @Tags questionnaires
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Questionnaire ID"
// @Param question body NewQuestionRequest true "Question data"
// @Success 201 {object} map[string]interface{} "Question created successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - only owner can create questions"
// @Failure 404 {object} map[string]string "Questionnaire not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/questionnaires/{id}/question [post]
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

	logger := GetLogger(c)
	logger.Info("Creating new question", zap.String("questionnaire_id", questionnaireID), zap.String("text", nq.Text), zap.String("theme", nq.Theme))
	nq.Sanitize()

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

// updateQuestionnaire updates an existing questionnaire (only if not published)
// @Summary Update questionnaire
// @Description Update a questionnaire's title and description (only owner and only if not published)
// @Tags questionnaires
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Questionnaire ID"
// @Param questionnaire body UpdateQuestionnaireRequest true "Updated questionnaire data"
// @Success 200 {object} map[string]interface{} "Questionnaire updated successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - only owner can update or questionnaire is published"
// @Failure 404 {object} map[string]string "Questionnaire not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/questionnaires/{id} [put]
func (s *Server) updateQuestionnaire(c echo.Context) error {
	entityIDStr, entityType, err := GetValuesFromToken(c)
	if err != nil || entityType != "user" {
		return c.JSON(401, map[string]string{"error": "unauthorized, invalid token"})
	}
	userID, err := uuid.Parse(entityIDStr)
	if err != nil {
		return c.JSON(401, map[string]string{"error": "unauthorized, invalid token"})
	}

	questionnaireID := c.Param("id")
	questionnaireUUID, err := uuid.Parse(questionnaireID)
	if err != nil {
		return c.JSON(400, map[string]string{"error": "invalid questionnaire ID"})
	}

	uq := new(UpdateQuestionnaireRequest)
	if err := BindAndValidate(c, uq); err != nil {
		return err
	}

	uq.Sanitize()

	ctx := c.Request().Context()

	q, err := s.service.GetQuestionnaire(questionnaireUUID, ctx)
	if err != nil {
		return c.JSON(404, map[string]string{"error": "questionnaire not found"})
	}

	u, err := q.QueryOwner().Only(ctx)
	if err != nil {
		return c.JSON(500, map[string]string{"error": "could not verify questionnaire owner"})
	}

	if u.ID != userID {
		return c.JSON(403, map[string]string{"error": "not authorized to update this questionnaire"})
	}

	if q.IsPublished {
		return c.JSON(403, map[string]string{"error": "cannot update published questionnaire"})
	}

	updatedQuestionnaire, err := s.service.UpdateQuestionnaire(questionnaireUUID, uq.Title, uq.Description, ctx)
	if err != nil {
		return c.JSON(500, map[string]string{"error": "could not update questionnaire"})
	}

	return c.JSON(200, updatedQuestionnaire)
}

// publishQuestionnaire publishes a questionnaire (makes it immutable)
// @Summary Publish questionnaire
// @Description Publish a questionnaire to make it available for responses (becomes immutable)
// @Tags questionnaires
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Questionnaire ID"
// @Success 200 {object} map[string]interface{} "Questionnaire published successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - only owner can publish"
// @Failure 404 {object} map[string]string "Questionnaire not found"
// @Failure 409 {object} map[string]string "Questionnaire already published"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/questionnaires/{id}/publish [post]
func (s *Server) publishQuestionnaire(c echo.Context) error {
	entityIDStr, entityType, err := GetValuesFromToken(c)
	if err != nil || entityType != "user" {
		return c.JSON(401, map[string]string{"error": "unauthorized, invalid token"})
	}
	userID, err := uuid.Parse(entityIDStr)
	if err != nil {
		return c.JSON(401, map[string]string{"error": "unauthorized, invalid token"})
	}

	questionnaireID := c.Param("id")
	questionnaireUUID, err := uuid.Parse(questionnaireID)
	if err != nil {
		return c.JSON(400, map[string]string{"error": "invalid questionnaire ID"})
	}

	ctx := c.Request().Context()

	q, err := s.service.GetQuestionnaire(questionnaireUUID, ctx)
	if err != nil {
		return c.JSON(404, map[string]string{"error": "questionnaire not found"})
	}

	u, err := q.QueryOwner().Only(ctx)
	if err != nil {
		return c.JSON(500, map[string]string{"error": "could not verify questionnaire owner"})
	}

	if u.ID != userID {
		return c.JSON(403, map[string]string{"error": "not authorized to publish this questionnaire"})
	}

	if q.IsPublished {
		return c.JSON(409, map[string]string{"error": "questionnaire already published"})
	}

	updatedQuestionnaire, err := s.service.PublishQuestionnaire(questionnaireUUID, userID, ctx)
	if err != nil {
		log := GetLogger(c)
		log.Error("failed to publish questionnaire",
			zap.String("questionnaire_id", questionnaireUUID.String()),
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return c.JSON(500, map[string]string{"error": "could not publish questionnaire"})
	}

	return c.JSON(200, updatedQuestionnaire)
}

// deleteQuestionnaire deletes a questionnaire and all related data (only if not published)
// @Summary Delete questionnaire
// @Description Delete a questionnaire and all its questions and members (only owner and only if not published)
// @Tags questionnaires
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Questionnaire ID"
// @Success 200 {object} map[string]string "Questionnaire deleted successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - only owner can delete or questionnaire is published"
// @Failure 404 {object} map[string]string "Questionnaire not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/questionnaires/{id} [delete]
func (s *Server) deleteQuestionnaire(c echo.Context) error {
	entityIDStr, entityType, err := GetValuesFromToken(c)
	if err != nil || entityType != "user" {
		return c.JSON(401, map[string]string{"error": "unauthorized, invalid token"})
	}
	userID, err := uuid.Parse(entityIDStr)
	if err != nil {
		return c.JSON(401, map[string]string{"error": "unauthorized, invalid token"})
	}

	questionnaireID := c.Param("id")
	questionnaireUUID, err := uuid.Parse(questionnaireID)
	if err != nil {
		return c.JSON(400, map[string]string{"error": "invalid questionnaire ID"})
	}

	ctx := c.Request().Context()

	q, err := s.service.GetQuestionnaire(questionnaireUUID, ctx)
	if err != nil {
		return c.JSON(404, map[string]string{"error": "questionnaire not found"})
	}

	u, err := q.QueryOwner().Only(ctx)
	if err != nil {
		return c.JSON(500, map[string]string{"error": "could not verify questionnaire owner"})
	}

	if u.ID != userID {
		return c.JSON(403, map[string]string{"error": "not authorized to delete this questionnaire"})
	}

	if q.IsPublished {
		return c.JSON(403, map[string]string{"error": "cannot delete published questionnaire"})
	}

	err = s.service.DeleteQuestionnaire(questionnaireUUID, ctx)
	if err != nil {
		return c.JSON(500, map[string]string{"error": "could not delete questionnaire"})
	}

	return c.JSON(200, map[string]string{"message": "questionnaire deleted successfully"})
}

// updateQuestion updates an existing question (only if questionnaire not published)
// @Summary Update question
// @Description Update a question's text and theme (only owner and only if questionnaire not published)
// @Tags questionnaires
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param questionnaireId path string true "Questionnaire ID"
// @Param questionId path string true "Question ID"
// @Param question body UpdateQuestionRequest true "Updated question data"
// @Success 200 {object} map[string]interface{} "Question updated successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - only owner can update or questionnaire is published"
// @Failure 404 {object} map[string]string "Question or questionnaire not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/questionnaires/{questionnaireId}/questions/{questionId} [put]
func (s *Server) updateQuestion(c echo.Context) error {
	entityIDStr, entityType, err := GetValuesFromToken(c)
	if err != nil || entityType != "user" {
		return c.JSON(401, map[string]string{"error": "unauthorized, invalid token"})
	}
	userID, err := uuid.Parse(entityIDStr)
	if err != nil {
		return c.JSON(401, map[string]string{"error": "unauthorized, invalid token"})
	}

	questionnaireID := c.Param("questionnaireId")
	questionID := c.Param("questionId")

	questionnaireUUID, err := uuid.Parse(questionnaireID)
	if err != nil {
		return c.JSON(400, map[string]string{"error": "invalid questionnaire ID"})
	}

	questionUUID, err := uuid.Parse(questionID)
	if err != nil {
		return c.JSON(400, map[string]string{"error": "invalid question ID"})
	}

	uq := new(UpdateQuestionRequest)
	if err := BindAndValidate(c, uq); err != nil {
		return err
	}

	uq.Sanitize()

	ctx := c.Request().Context()

	q, err := s.service.GetQuestionnaire(questionnaireUUID, ctx)
	if err != nil {
		return c.JSON(404, map[string]string{"error": "questionnaire not found"})
	}

	u, err := q.QueryOwner().Only(ctx)
	if err != nil {
		return c.JSON(500, map[string]string{"error": "could not verify questionnaire owner"})
	}

	if u.ID != userID {
		return c.JSON(403, map[string]string{"error": "not authorized to update questions in this questionnaire"})
	}

	if q.IsPublished {
		return c.JSON(403, map[string]string{"error": "cannot update questions in published questionnaire"})
	}

	question, err := s.service.GetQuestionWithQuestionnaire(questionUUID, ctx)
	if err != nil {
		return c.JSON(404, map[string]string{"error": "question not found"})
	}

	if question.Edges.Questionnaire.ID != questionnaireUUID {
		return c.JSON(400, map[string]string{"error": "question does not belong to specified questionnaire"})
	}

	updatedQuestion, err := s.service.UpdateQuestion(questionUUID, uq.Text, uq.Theme, ctx)
	if err != nil {
		return c.JSON(500, map[string]string{"error": "could not update question"})
	}

	return c.JSON(200, updatedQuestion)
}

// deleteQuestion deletes a question (only if questionnaire not published)
// @Summary Delete question
// @Description Delete a question from a questionnaire (only owner and only if questionnaire not published)
// @Tags questionnaires
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param questionnaireId path string true "Questionnaire ID"
// @Param questionId path string true "Question ID"
// @Success 200 {object} map[string]string "Question deleted successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - only owner can delete or questionnaire is published"
// @Failure 404 {object} map[string]string "Question or questionnaire not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/questionnaires/{questionnaireId}/questions/{questionId} [delete]
func (s *Server) deleteQuestion(c echo.Context) error {
	entityIDStr, entityType, err := GetValuesFromToken(c)
	if err != nil || entityType != "user" {
		return c.JSON(401, map[string]string{"error": "unauthorized, invalid token"})
	}
	userID, err := uuid.Parse(entityIDStr)
	if err != nil {
		return c.JSON(401, map[string]string{"error": "unauthorized, invalid token"})
	}

	questionnaireID := c.Param("questionnaireId")
	questionID := c.Param("questionId")

	questionnaireUUID, err := uuid.Parse(questionnaireID)
	if err != nil {
		return c.JSON(400, map[string]string{"error": "invalid questionnaire ID"})
	}

	questionUUID, err := uuid.Parse(questionID)
	if err != nil {
		return c.JSON(400, map[string]string{"error": "invalid question ID"})
	}

	ctx := c.Request().Context()

	q, err := s.service.GetQuestionnaire(questionnaireUUID, ctx)
	if err != nil {
		return c.JSON(404, map[string]string{"error": "questionnaire not found"})
	}

	u, err := q.QueryOwner().Only(ctx)
	if err != nil {
		return c.JSON(500, map[string]string{"error": "could not verify questionnaire owner"})
	}

	if u.ID != userID {
		return c.JSON(403, map[string]string{"error": "not authorized to delete questions in this questionnaire"})
	}

	if q.IsPublished {
		return c.JSON(403, map[string]string{"error": "cannot delete questions in published questionnaire"})
	}

	question, err := s.service.GetQuestionWithQuestionnaire(questionUUID, ctx)
	if err != nil {
		return c.JSON(404, map[string]string{"error": "question not found"})
	}

	if question.Edges.Questionnaire.ID != questionnaireUUID {
		return c.JSON(400, map[string]string{"error": "question does not belong to specified questionnaire"})
	}

	err = s.service.DeleteQuestion(questionUUID, ctx)
	if err != nil {
		return c.JSON(500, map[string]string{"error": "could not delete question"})
	}

	return c.JSON(200, map[string]string{"message": "question deleted successfully"})
}

// getUserQuestionnaires returns all questionnaires owned by the authenticated user
// @Summary Get user questionnaires
// @Description Get all questionnaires owned by the authenticated user
// @Tags questionnaires
// @Produce json
// @Security BearerAuth
// @Success 200 {array} object "List of questionnaires"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/questionnaires [get]
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
// @Summary Get questionnaire details
// @Description Get detailed information about a questionnaire including questions and answers (if owner) or just questions (if member)
// @Tags questionnaires
// @Produce json
// @Security BearerAuth
// @Param id path string true "Questionnaire ID"
// @Success 200 {object} object "Questionnaire details"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Not found"
// @Router /api/questionnaires/{id} [get]
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
// @Summary Get questionnaire questions
// @Description Get all questions for a specific questionnaire
// @Tags questionnaires
// @Produce json
// @Security BearerAuth
// @Param id path string true "Questionnaire ID"
// @Success 200 {array} object "List of questions"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Not found"
// @Router /api/questionnaires/{id}/questions [get]
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
	questionnaire, err := s.service.GetQuestionnaireWithDetails(qID, ctx)
	if err != nil {
		return c.JSON(404, map[string]string{"error": "questionnaire not found"})
	}

	userID, _ := uuid.Parse(entityIDStr)
	isOwner := entityType == "user" && questionnaire.Edges.Owner.ID == userID
	isMember := false

	switch entityType {
	case "user":
		_, err := s.service.GetMemberByUserAndQuestionnaire(userID, qID, ctx)
		isMember = err == nil
	case "member":
		member, err := s.service.GetMemberWithQuestionnaire(userID, ctx)
		isMember = err == nil && member.Edges.Questionnaire.ID == qID
	}

	if !isOwner && !isMember {
		return c.JSON(403, map[string]string{"error": "forbidden"})
	}

	questions, err := s.service.GetQuestionnaireQuestionsWithAnswers(qID, ctx)
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
// @Summary Get member answers
// @Description Get all answers provided by the authenticated user/member for a specific questionnaire
// @Tags questionnaires
// @Produce json
// @Security BearerAuth
// @Param id path string true "Questionnaire ID"
// @Success 200 {array} object "List of answers"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Not found"
// @Router /api/questionnaires/{id}/my-answers [get]
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
