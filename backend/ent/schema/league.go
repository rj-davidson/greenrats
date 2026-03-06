package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type League struct {
	ent.Schema
}

func (League) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		OwnershipMixin{},
	}
}

func (League) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			MaxLen(50),
		field.String("code").
			Unique().
			NotEmpty().
			Comment("Shareable join code"),
		field.Int("season_year"),
		field.Bool("joining_enabled").
			Default(true),
	}
}

func (League) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("memberships", LeagueMembership.Type),
		edge.To("picks", Pick.Type),
		edge.To("commissioner_actions", CommissionerAction.Type),
		edge.To("email_reminders", EmailReminder.Type),
		edge.From("season", Season.Type).
			Ref("leagues").
			Unique().
			Required(),
	}
}
