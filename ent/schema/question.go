package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Question holds the schema definition for the Question entity.
type Question struct {
	ent.Schema
}

// Fields of the Question.
func (Question) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()),
		field.String("text"),
	}
}

// Edges of the Question.
func (Question) Edges() []ent.Edge {
	return nil
}
