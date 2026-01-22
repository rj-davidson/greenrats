package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// HoleScore holds the schema definition for the HoleScore entity.
type HoleScore struct {
	ent.Schema
}

// Mixin of the HoleScore.
func (HoleScore) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
	}
}

// Fields of the HoleScore.
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

// Edges of the HoleScore.
func (HoleScore) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("round", Round.Type).
			Ref("hole_scores").
			Unique().
			Required(),
	}
}

// Indexes of the HoleScore.
func (HoleScore) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("round").
			Fields("hole_number").
			Unique(),
	}
}
