package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// League holds the schema definition for the League entity.
type League struct {
	ent.Schema
}

// Mixin of the League.
func (League) Mixin() []ent.Mixin {
	return []ent.Mixin{
		OwnershipMixin{},
	}
}

// Fields of the League.
func (League) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty(),
		field.String("code").
			Unique().
			NotEmpty(),
		field.Int("season_year"),
	}
}

// Edges of the League.
func (League) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("memberships", LeagueMembership.Type),
		edge.To("picks", Pick.Type),
	}
}
