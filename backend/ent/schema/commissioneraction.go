package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type CommissionerAction struct {
	ent.Schema
}

func (CommissionerAction) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
	}
}

func (CommissionerAction) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("action_type").
			Values("pick_change", "join_code_reset", "joining_disabled", "joining_enabled", "member_removed"),
		field.String("description").
			NotEmpty(),
		field.JSON("metadata", map[string]any{}).
			Optional().
			SchemaType(map[string]string{
				dialect.Postgres: "jsonb",
			}),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

func (CommissionerAction) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("league", League.Type).
			Ref("commissioner_actions").
			Unique().
			Required(),
		edge.From("commissioner", User.Type).
			Ref("commissioner_actions").
			Unique().
			Required(),
		edge.From("affected_user", User.Type).
			Ref("affected_actions").
			Unique(),
	}
}
