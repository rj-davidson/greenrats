package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Placement struct {
	ent.Schema
}

func (Placement) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		BaseMixin{},
	}
}

func (Placement) Fields() []ent.Field {
	return []ent.Field{
		field.String("position").
			Default("").
			Comment("Position from API: '1', 'T5', 'CUT', 'WD'"),
		field.Int("position_numeric").
			Optional().
			Nillable().
			Comment("Numeric position for sorting (null for CUT/WD)"),
		field.Int("total_score").
			Optional().
			Nillable().
			Comment("Total strokes (e.g., 257)"),
		field.Int("par_relative_score").
			Optional().
			Nillable().
			Comment("Score relative to par (e.g., -35)"),
		field.Int("earnings").
			Default(0).
			Comment("Prize money in dollars"),
		field.Enum("status").
			Values("cut", "withdrawn", "finished").
			Default("finished").
			Comment("Golfer's final status in the tournament"),
	}
}

func (Placement) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tournament", Tournament.Type).
			Ref("placements").
			Unique().
			Required(),
		edge.From("golfer", Golfer.Type).
			Ref("placements").
			Unique().
			Required(),
	}
}

func (Placement) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("tournament", "golfer").
			Unique(),
	}
}
