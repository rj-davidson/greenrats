package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Round holds the schema definition for the Round entity.
type Round struct {
	ent.Schema
}

// Mixin of the Round.
func (Round) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		BaseMixin{},
	}
}

// Fields of the Round.
func (Round) Fields() []ent.Field {
	return []ent.Field{
		field.Int("round_number").
			Range(1, 4),
		field.Int("score").
			Optional().
			Nillable().
			Comment("Total strokes for round (e.g., 68)"),
		field.Int("par_relative_score").
			Optional().
			Nillable().
			Comment("Score relative to par (e.g., -4)"),
		field.Time("tee_time").
			Optional().
			Nillable(),
	}
}

// Edges of the Round.
func (Round) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("leaderboard_entry", LeaderboardEntry.Type).
			Ref("rounds").
			Unique().
			Required(),
		edge.To("hole_scores", HoleScore.Type),
		edge.From("course", Course.Type).
			Ref("rounds").
			Unique(),
	}
}

// Indexes of the Round.
func (Round) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("leaderboard_entry").
			Fields("round_number").
			Unique(),
	}
}
