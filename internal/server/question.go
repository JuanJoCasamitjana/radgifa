package server

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type AnswerRequest struct {
	AnswerValue string `json:"answer_value" validate:"required,oneof=Yes No Pass" example:"Yes"`
}

func (s *Server) newQuestionAnswer(c echo.Context) error {
	log := GetLogger(c)

	questionIDStr := c.Param("id")
	questionID, err := uuid.Parse(questionIDStr)
	if err != nil {
		log.Error("Invalid question ID", zap.String("question_id", questionIDStr), zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid question ID format",
		})
	}

	entityID, entityType, err := GetValuesFromToken(c)
	if err != nil {
		log.Error("Failed to get values from token", zap.Error(err))
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid or missing token",
		})
	}

	var req AnswerRequest
	if err := c.Bind(&req); err != nil {
		log.Error("Failed to bind request", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	if err := c.Validate(&req); err != nil {
		log.Error("Request validation failed", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Validation failed: " + err.Error(),
		})
	}

	ctx := c.Request().Context()

	question, err := s.service.GetQuestionWithQuestionnaire(questionID, ctx)
	if err != nil {
		log.Error("Failed to get question", zap.String("question_id", questionIDStr), zap.Error(err))
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Question not found",
		})
	}

	var memberID uuid.UUID

	switch entityType {
	case "user":
		userID, err := uuid.Parse(entityID)
		if err != nil {
			log.Error("Invalid user ID from token", zap.String("entity_id", entityID), zap.Error(err))
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid user ID",
			})
		}

		member, err := s.service.GetMemberByUserAndQuestionnaire(userID, question.Edges.Questionnaire.ID, ctx)
		if err != nil {
			log.Error("User is not a member of this questionnaire",
				zap.String("user_id", userID.String()),
				zap.String("questionnaire_id", question.Edges.Questionnaire.ID.String()),
				zap.Error(err))
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "You are not a member of this questionnaire",
			})
		}
		memberID = member.ID

	case "member":
		memberIDParsed, err := uuid.Parse(entityID)
		if err != nil {
			log.Error("Invalid member ID from token", zap.String("entity_id", entityID), zap.Error(err))
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid member ID",
			})
		}

		member, err := s.service.GetMemberWithQuestionnaire(memberIDParsed, ctx)
		if err != nil {
			log.Error("Failed to get member", zap.String("member_id", entityID), zap.Error(err))
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Member not found",
			})
		}

		if member.Edges.Questionnaire.ID != question.Edges.Questionnaire.ID {
			log.Error("Member does not belong to question's questionnaire",
				zap.String("member_questionnaire_id", member.Edges.Questionnaire.ID.String()),
				zap.String("question_questionnaire_id", question.Edges.Questionnaire.ID.String()))
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "You are not authorized to answer this question",
			})
		}
		memberID = member.ID

	default:
		log.Error("Invalid entity type", zap.String("entity_type", entityType))
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid token entity type",
		})
	}

	answer, err := s.service.CreateAnswer(memberID, questionID, req.AnswerValue, ctx)
	if err != nil {
		log.Error("Failed to create answer",
			zap.String("member_id", memberID.String()),
			zap.String("question_id", questionID.String()),
			zap.String("answer_value", req.AnswerValue),
			zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create answer",
		})
	}

	log.Info("Answer created successfully",
		zap.String("answer_id", answer.ID.String()),
		zap.String("member_id", memberID.String()),
		zap.String("question_id", questionID.String()),
		zap.String("answer_value", req.AnswerValue))

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message":      "Answer created successfully",
		"answer_id":    answer.ID,
		"question_id":  questionID,
		"member_id":    memberID,
		"answer_value": answer.AnswerValue,
		"created_at":   answer.CreatedAt,
	})
}
