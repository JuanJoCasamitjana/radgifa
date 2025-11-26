package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Answer holds the schema definition for the Answer entity.
type Answer struct {
	ent.Schema
}

// Fields of the Answer.
func (Answer) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()),
		field.Enum("answer_value").Values("Yes", "No", "Pass"),
	}
}

// Edges of the Answer.
func (Answer) Edges() []ent.Edge {
	return nil
}
