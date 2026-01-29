package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Round struct {
	ent.Schema
}

func (Round) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		BaseMixin{},
	}
}

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
		field.Int("thru").
			Optional().
			Nillable().
			Comment("Number of holes completed (0-18)"),
	}
}

func (Round) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tournament", Tournament.Type).
			Ref("rounds").
			Unique().
			Required(),
		edge.From("golfer", Golfer.Type).
			Ref("rounds").
			Unique().
			Required(),
		edge.To("hole_scores", HoleScore.Type),
		edge.From("course", Course.Type).
			Ref("rounds").
			Unique(),
	}
}

func (Round) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("tournament", "golfer").
			Fields("round_number").
			Unique(),
	}
}
