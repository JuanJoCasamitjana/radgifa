package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Questionnaire holds the schema definition for the Questionnaire entity.
type Questionnaire struct {
	ent.Schema
}

// Fields of the Questionnaire.
func (Questionnaire) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()),
		field.String("title"),
		field.String("description").Optional(),
		field.Bool("is_published").Default(false),
		field.Int("created_at").DefaultFunc(time.Now().UnixMilli()).Immutable(),
	}
}

// Edges of the Questionnaire.
func (Questionnaire) Edges() []ent.Edge {
	return nil
}
