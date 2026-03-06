package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Season struct {
	ent.Schema
}

func (Season) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		BaseMixin{},
	}
}

func (Season) Fields() []ent.Field {
	return []ent.Field{
		field.Int("year").
			Unique().
			Positive(),
		field.Time("start_date"),
		field.Time("end_date"),
		field.Bool("is_current").
			Default(false),
	}
}

func (Season) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("tournaments", Tournament.Type),
		edge.To("leagues", League.Type),
		edge.To("picks", Pick.Type),
		edge.To("golfer_seasons", GolferSeason.Type),
	}
}
