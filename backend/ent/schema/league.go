package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// League holds the schema definition for the League entity.
type League struct {
	ent.Schema
}

// Mixin of the League.
func (League) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		OwnershipMixin{},
	}
}

// Fields of the League.
func (League) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty(),
		field.String("code").
			Unique().
			NotEmpty().
			Comment("Shareable join code"),
		field.Int("season_year"),
	}
}

// Edges of the League.
func (League) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("memberships", LeagueMembership.Type),
		edge.To("picks", Pick.Type),
		edge.To("commissioner_actions", CommissionerAction.Type),
	}
}
