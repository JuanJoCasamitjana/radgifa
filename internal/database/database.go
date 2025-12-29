package database

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"radgifa/ent"
	"radgifa/ent/answer"
	"radgifa/ent/member"
	"radgifa/ent/question"
	"radgifa/ent/questionnaire"
	"radgifa/ent/user"

	"entgo.io/ent/dialect"
	enSQL "entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/crypto/bcrypt"
)

var (
	bcryptCost = getCurrentBcryptCost()
)

type Service interface {
	Health() map[string]string
	Close() error
	Client() *ent.Client

	CreateUser(name, displayName, username, password string, ctx context.Context) (*ent.User, error)
	ValidateUserCredentials(username, password string, ctx context.Context) (*ent.User, error)
	IsUsernameAvailable(username string, ctx context.Context) (bool, error)
	CreateQuestionnaire(userID uuid.UUID, title, description string, ctx context.Context) (*ent.Questionnaire, error)
	GetQuestionnaire(questionnaireID uuid.UUID, ctx context.Context) (*ent.Questionnaire, error)

	CreateMember(userID, questionnaireID uuid.UUID, uniqueIdentifier, displayName string, ctx context.Context) (*ent.Member, error)
	CreateAnonymousMember(questionnaireID uuid.UUID, uniqueIdentifier, displayName string, ctx context.Context) (*ent.Member, string, error)
	ValidateMemberCredentials(uniqueIdentifier, passcode string, ctx context.Context) (*ent.Member, error)
	GetMemberWithQuestionnaire(memberID uuid.UUID, ctx context.Context) (*ent.Member, error)
	IsMemberIdentifierAvailable(questionnaireID uuid.UUID, uniqueIdentifier string, ctx context.Context) (bool, error)
	CreateNewQuestion(questionnaireID uuid.UUID, text, theme string, ctx context.Context) (*ent.Question, error)
	GetQuestionWithQuestionnaire(questionID uuid.UUID, ctx context.Context) (*ent.Question, error)
	GetMemberByUserAndQuestionnaire(userID, questionnaireID uuid.UUID, ctx context.Context) (*ent.Member, error)
	CreateAnswer(memberID, questionID uuid.UUID, answerValue string, ctx context.Context) (*ent.Answer, error)

	// New GET methods
	GetUserQuestionnaires(userID uuid.UUID, ctx context.Context) ([]*ent.Questionnaire, error)
	GetQuestionnaireWithDetails(questionnaireID uuid.UUID, ctx context.Context) (*ent.Questionnaire, error)
	GetQuestionnaireQuestions(questionnaireID uuid.UUID, ctx context.Context) ([]*ent.Question, error)
	GetQuestionnaireMembers(questionnaireID uuid.UUID, ctx context.Context) ([]*ent.Member, error)
	GetMemberAnswers(memberID, questionnaireID uuid.UUID, ctx context.Context) ([]*ent.Answer, error)
}

type service struct {
	db     *sql.DB
	client *ent.Client
}

var (
	database   = os.Getenv("DB_NAME")
	password   = os.Getenv("DB_PASSWORD")
	username   = os.Getenv("DB_USERNAME")
	port       = os.Getenv("DB_PORT")
	host       = os.Getenv("DB_HOST")
	schema     = os.Getenv("DB_SCHEMA")
	dbInstance *service
)

func New() Service {
	if dbInstance != nil {
		return dbInstance
	}
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s", username, password, host, port, database, schema)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}

	drv := enSQL.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(drv))

	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	dbInstance = &service{
		db:     db,
		client: client,
	}
	return dbInstance
}

func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf("db down: %v", err)
		return stats
	}

	stats["status"] = "up"
	stats["message"] = "It's healthy"

	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	if dbStats.OpenConnections > 40 {
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", database)
	if err := s.client.Close(); err != nil {
		log.Printf("Error closing Ent client: %v", err)
	}
	return s.db.Close()
}

// Client returns the Ent client instance.
func (s *service) Client() *ent.Client {
	return s.client
}

func getCurrentBcryptCost() int {
	costStr := os.Getenv("BCRYPT_COST")
	if costStr == "" {
		return bcrypt.DefaultCost
	}
	cost, err := strconv.Atoi(costStr)
	if err != nil {
		return bcrypt.DefaultCost
	}
	return cost
}

func generateSecurePasscode() (string, error) {
	bytes := make([]byte, 6) // 6 bytes = 48 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:8], nil
}

func (s *service) CreateUser(name, displayName, username, password string, ctx context.Context) (*ent.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, err
	}
	user, err := s.client.User.Create().SetName(name).SetDisplayName(displayName).SetUsername(username).SetPassword(hashedPassword).Save(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *service) ValidateUserCredentials(username, password string, ctx context.Context) (*ent.User, error) {
	user, err := s.client.User.Query().Where(user.Username(username)).First(ctx)
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword(user.Password, []byte(password))
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *service) IsUsernameAvailable(username string, ctx context.Context) (bool, error) {
	count, err := s.client.User.Query().
		Where(user.Username(username)).
		Count(ctx)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func (s *service) CreateQuestionnaire(userID uuid.UUID, title, description string, ctx context.Context) (*ent.Questionnaire, error) {
	questionnaire, err := s.client.Questionnaire.Create().
		SetTitle(title).
		SetDescription(description).
		SetOwnerID(userID).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return questionnaire, nil
}

func (s *service) GetQuestionnaire(questionnaireID uuid.UUID, ctx context.Context) (*ent.Questionnaire, error) {
	questionnaire, err := s.client.Questionnaire.
		Query().
		Where(questionnaire.ID(questionnaireID)).
		WithOwner().
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return questionnaire, nil
}

func (s *service) CreateMember(userID, questionnaireID uuid.UUID, uniqueIdentifier, displayName string, ctx context.Context) (*ent.Member, error) {
	member, err := s.client.Member.Create().
		SetQuestionnaireID(questionnaireID).
		SetUserID(userID).
		SetUniqueIdentifier(uniqueIdentifier).
		SetDisplayName(displayName).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return member, nil
}

func (s *service) CreateAnonymousMember(questionnaireID uuid.UUID, uniqueIdentifier, displayName string, ctx context.Context) (*ent.Member, string, error) {
	passcode, err := generateSecurePasscode()
	if err != nil {
		return nil, "", err
	}

	hashedPasscode, err := bcrypt.GenerateFromPassword([]byte(passcode), bcryptCost)
	if err != nil {
		return nil, "", err
	}

	member, err := s.client.Member.Create().
		SetQuestionnaireID(questionnaireID).
		SetDisplayName(displayName).
		SetUniqueIdentifier(uniqueIdentifier).
		SetPassCode(hashedPasscode).
		Save(ctx)
	if err != nil {
		return nil, "", err
	}
	return member, passcode, nil
}

func (s *service) ValidateMemberCredentials(uniqueIdentifier, passcode string, ctx context.Context) (*ent.Member, error) {
	member, err := s.client.Member.Query().
		Where(member.UniqueIdentifier(uniqueIdentifier)).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword(member.PassCode, []byte(passcode))
	if err != nil {
		return nil, err
	}
	return member, nil
}

func (s *service) GetMemberWithQuestionnaire(memberID uuid.UUID, ctx context.Context) (*ent.Member, error) {
	return s.client.Member.Query().
		Where(member.ID(memberID)).
		WithQuestionnaire().
		Only(ctx)
}

func (s *service) IsMemberIdentifierAvailable(questionnaireID uuid.UUID, uniqueIdentifier string, ctx context.Context) (bool, error) {
	count, err := s.client.Member.Query().
		Where(
			member.HasQuestionnaireWith(questionnaire.ID(questionnaireID)),
			member.UniqueIdentifier(uniqueIdentifier),
		).
		Count(ctx)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func (s *service) CreateNewQuestion(questionnaireID uuid.UUID, text, theme string, ctx context.Context) (*ent.Question, error) {
	question, err := s.client.Question.Create().
		SetQuestionnaireID(questionnaireID).
		SetText(text).
		SetTheme(theme).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return question, nil
}

func (s *service) GetQuestionWithQuestionnaire(questionID uuid.UUID, ctx context.Context) (*ent.Question, error) {
	return s.client.Question.Query().
		Where(question.ID(questionID)).
		WithQuestionnaire().
		Only(ctx)
}

func (s *service) GetMemberByUserAndQuestionnaire(userID, questionnaireID uuid.UUID, ctx context.Context) (*ent.Member, error) {
	return s.client.Member.Query().
		Where(
			member.HasUserWith(user.ID(userID)),
			member.HasQuestionnaireWith(questionnaire.ID(questionnaireID)),
		).
		Only(ctx)
}

func (s *service) CreateAnswer(memberID, questionID uuid.UUID, answerValue string, ctx context.Context) (*ent.Answer, error) {
	existingAnswer, err := s.client.Answer.Query().
		Where(
			answer.HasMemberWith(member.ID(memberID)),
			answer.HasQuestionWith(question.ID(questionID)),
		).Only(ctx)

	if err == nil {
		return existingAnswer.Update().
			SetAnswerValue(answer.AnswerValue(answerValue)).
			SetUpdatedAt(time.Now().UnixMilli()).
			Save(ctx)
	}

	if !ent.IsNotFound(err) {
		return nil, err
	}

	return s.client.Answer.Create().
		SetMemberID(memberID).
		SetQuestionID(questionID).
		SetAnswerValue(answer.AnswerValue(answerValue)).
		Save(ctx)
}

// GetUserQuestionnaires returns all questionnaires owned by a user
func (s *service) GetUserQuestionnaires(userID uuid.UUID, ctx context.Context) ([]*ent.Questionnaire, error) {
	return s.client.Questionnaire.Query().
		Where(questionnaire.HasOwnerWith(user.ID(userID))).
		Order(ent.Desc("created_at")).
		All(ctx)
}

// GetQuestionnaireWithDetails returns a questionnaire with all its relations
func (s *service) GetQuestionnaireWithDetails(questionnaireID uuid.UUID, ctx context.Context) (*ent.Questionnaire, error) {
	return s.client.Questionnaire.Query().
		Where(questionnaire.ID(questionnaireID)).
		WithOwner().
		WithMembers(func(q *ent.MemberQuery) {
			q.WithUser()
		}).
		WithQuestions(func(q *ent.QuestionQuery) {
			q.WithAnswers(func(a *ent.AnswerQuery) {
				a.WithMember()
			})
		}).
		Only(ctx)
}

// GetQuestionnaireQuestions returns all questions for a questionnaire
func (s *service) GetQuestionnaireQuestions(questionnaireID uuid.UUID, ctx context.Context) ([]*ent.Question, error) {
	return s.client.Question.Query().
		Where(question.HasQuestionnaireWith(questionnaire.ID(questionnaireID))).
		Order(ent.Asc("created_at")).
		All(ctx)
}

// GetQuestionnaireMembers returns all members of a questionnaire
func (s *service) GetQuestionnaireMembers(questionnaireID uuid.UUID, ctx context.Context) ([]*ent.Member, error) {
	return s.client.Member.Query().
		Where(member.HasQuestionnaireWith(questionnaire.ID(questionnaireID))).
		WithUser().
		Order(ent.Asc("created_at")).
		All(ctx)
}

// GetMemberAnswers returns all answers by a member for a specific questionnaire
func (s *service) GetMemberAnswers(memberID, questionnaireID uuid.UUID, ctx context.Context) ([]*ent.Answer, error) {
	return s.client.Answer.Query().
		Where(
			answer.HasMemberWith(member.ID(memberID)),
			answer.HasQuestionWith(
				question.HasQuestionnaireWith(questionnaire.ID(questionnaireID)),
			),
		).
		WithQuestion().
		Order(ent.Asc("created_at")).
		All(ctx)
}
