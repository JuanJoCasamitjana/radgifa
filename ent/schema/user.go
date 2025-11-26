package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()),
		field.String("name"),
		field.String("display_name").Optional(),
		field.Int("created_at").DefaultFunc(time.Now().UnixMilli()).Immutable(),
		field.Bytes("password"),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("questionnaire", Questionnaire.Type),
		edge.To("memberships", Member.Type),
	}
}
