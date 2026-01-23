package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Mixin of the User.
func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		BaseMixin{},
	}
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("workos_id").
			Unique().
			NotEmpty(),
		field.String("email").
			NotEmpty(),
		field.String("display_name").
			Unique().
			Optional().
			Nillable().
			MaxLen(20),
		field.Bool("is_admin").
			Default(false),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("picks", Pick.Type),
		edge.To("league_memberships", LeagueMembership.Type),
		edge.To("commissioner_actions", CommissionerAction.Type),
		edge.To("affected_actions", CommissionerAction.Type),
		edge.To("email_reminders", EmailReminder.Type),
	}
}
