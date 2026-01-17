package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// LeagueMembership holds the schema definition for the LeagueMembership entity.
type LeagueMembership struct {
	ent.Schema
}

// Mixin of the LeagueMembership.
func (LeagueMembership) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		BaseMixin{},
	}
}

// Fields of the LeagueMembership.
func (LeagueMembership) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("role").
			Values("owner", "member").
			Default("member"),
		field.Time("joined_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the LeagueMembership.
func (LeagueMembership) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("league_memberships").
			Unique().
			Required(),
		edge.From("league", League.Type).
			Ref("memberships").
			Unique().
			Required(),
	}
}

// Indexes of the LeagueMembership.
func (LeagueMembership) Indexes() []ent.Index {
	return []ent.Index{
		// One membership per user per league
		index.Edges("user", "league").
			Unique(),
	}
}
