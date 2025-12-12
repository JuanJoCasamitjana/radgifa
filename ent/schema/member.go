package schema

import (
	"time"

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
		field.UUID("id", uuid.New()).Default(uuid.New).Immutable(),
		field.String("display_name"),
		field.Int64("created_at").DefaultFunc(func() int64 { return time.Now().UnixMilli() }).Immutable(),
		field.String("unique_identifier").Comment("Something that only the member knows so they can prove who they are"),
		field.Bytes("pass_code").Comment("It is generated as a string the clear text is send to the member only once, then only the hash is stored. Is not requiered if the member is related to a user."),
	}
}

// Edges of the Member.
func (Member) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("memberships").Unique(),
		edge.From("questionnaire", Questionnaire.Type).Ref("members").Unique().Required(),
		edge.To("answers", Answer.Type),
	}
}
