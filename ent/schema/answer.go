package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
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
		field.Int64("created_at").DefaultFunc(func() int64 { return time.Now().UnixMilli() }).Immutable(),
		field.Int64("updated_at").DefaultFunc(func() int64 { return time.Now().UnixMilli() }),
	}
}

// Edges of the Answer.
func (Answer) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("question", Question.Type).Ref("answers").Unique().Required(),
		edge.From("member", Member.Type).Ref("answers").Unique().Required(),
	}
}
