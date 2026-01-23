package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Pick holds the schema definition for the Pick entity.
type Pick struct {
	ent.Schema
}

// Mixin of the Pick.
func (Pick) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
	}
}

// Fields of the Pick.
func (Pick) Fields() []ent.Field {
	return []ent.Field{
		field.Int("season_year"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the Pick.
func (Pick) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("picks").
			Unique().
			Required(),
		edge.From("tournament", Tournament.Type).
			Ref("picks").
			Unique().
			Required(),
		edge.From("golfer", Golfer.Type).
			Ref("picks").
			Unique().
			Required(),
		edge.From("league", League.Type).
			Ref("picks").
			Unique().
			Required(),
		edge.From("season", Season.Type).
			Ref("picks").
			Unique().
			Required(),
	}
}

// Indexes of the Pick.
func (Pick) Indexes() []ent.Index {
	return []ent.Index{
		// One pick per user per tournament per league
		index.Edges("user", "tournament", "league").
			Unique(),
		// Cannot reuse a golfer within a league-year
		index.Edges("user", "golfer", "league").
			Fields("season_year").
			Unique(),
	}
}
