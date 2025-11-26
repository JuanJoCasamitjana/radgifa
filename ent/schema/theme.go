package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Theme holds the schema definition for the Theme entity.
type Theme struct {
	ent.Schema
}

// Fields of the Theme.
func (Theme) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()),
		field.String("name"),
	}
}

// Edges of the Theme.
func (Theme) Edges() []ent.Edge {
	return nil
}
