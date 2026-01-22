package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// CourseHole holds the schema definition for the CourseHole entity.
type CourseHole struct {
	ent.Schema
}

// Mixin of the CourseHole.
func (CourseHole) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
	}
}

// Fields of the CourseHole.
func (CourseHole) Fields() []ent.Field {
	return []ent.Field{
		field.Int("hole_number").
			Range(1, 18),
		field.Int("par").
			Range(3, 5),
		field.Int("yardage").
			Optional().
			Nillable(),
	}
}

// Edges of the CourseHole.
func (CourseHole) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("course", Course.Type).
			Ref("holes").
			Unique().
			Required(),
	}
}

// Indexes of the CourseHole.
func (CourseHole) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("course").
			Fields("hole_number").
			Unique(),
	}
}
