package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Course holds the schema definition for the Course entity.
type Course struct {
	ent.Schema
}

// Mixin of the Course.
func (Course) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		BaseMixin{},
	}
}

// Fields of the Course.
func (Course) Fields() []ent.Field {
	return []ent.Field{
		field.Int("bdl_id").
			Optional().
			Nillable().
			Unique().
			Comment("BallDontLie API ID"),
		field.String("pga_tour_id").
			Optional().
			Nillable().
			Unique().
			Comment("PGA Tour ID (for future scraping)"),
		field.String("name").
			NotEmpty(),
		field.Int("par").
			Optional().
			Nillable(),
		field.Int("yardage").
			Optional().
			Nillable(),
		field.String("city").
			Optional().
			Nillable(),
		field.String("state").
			Optional().
			Nillable(),
		field.String("country").
			Optional().
			Nillable(),
	}
}

// Edges of the Course.
func (Course) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("holes", CourseHole.Type),
		edge.To("tournaments", Tournament.Type),
	}
}
