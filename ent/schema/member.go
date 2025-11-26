package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Member holds the schema definition for the Member entity.
type Member struct {
	ent.Schema
}

// Fields of the Member.
func (Member) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()),
		field.String("display_name"),
		field.String("unique_identifier").Comment("Something that only the member knows so they can prove who they are"),
		field.Bytes("pass_code").Comment("It is generated as a string the clear text is send to the member only once, then only the hash is stored"),
	}
}

// Edges of the Member.
func (Member) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Unique(),
		edge.To("questionnaire", Questionnaire.Type).Unique(),
		edge.To("answers", Answer.Type),
	}
}
