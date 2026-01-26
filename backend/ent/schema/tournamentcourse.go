package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type TournamentCourse struct {
	ent.Schema
}

func (TournamentCourse) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		BaseMixin{},
	}
}

func (TournamentCourse) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("rounds", []int{}).
			Optional().
			Comment("Which round numbers use this course (e.g., [1,2] or [1,2,3,4])"),
	}
}

func (TournamentCourse) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tournament", Tournament.Type).
			Ref("tournament_courses").
			Unique().
			Required(),
		edge.From("course", Course.Type).
			Ref("tournament_courses").
			Unique().
			Required(),
	}
}

func (TournamentCourse) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("tournament", "course").Unique(),
	}
}
