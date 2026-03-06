package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type HoleScore struct {
	ent.Schema
}

func (HoleScore) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
	}
}

func (HoleScore) Fields() []ent.Field {
	return []ent.Field{
		field.Int("hole_number").
			Range(1, 18),
		field.Int("par").
			Range(3, 5),
		field.Int("score").
			Optional().
			Nillable().
			Comment("Player's score on this hole"),
	}
}

func (HoleScore) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("round", Round.Type).
			Ref("hole_scores").
			Unique().
			Required(),
	}
}

func (HoleScore) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("round").
			Fields("hole_number").
			Unique(),
	}
}
