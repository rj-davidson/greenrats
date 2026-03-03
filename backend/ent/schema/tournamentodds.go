package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type TournamentOdds struct {
	ent.Schema
}

func (TournamentOdds) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		BaseMixin{},
	}
}

func (TournamentOdds) Fields() []ent.Field {
	return []ent.Field{
		field.String("vendor"),
		field.Int("american_odds"),
		field.Float("implied_probability"),
		field.Time("odds_updated_at"),
	}
}

func (TournamentOdds) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tournament", Tournament.Type).
			Ref("tournament_odds").
			Unique().
			Required(),
		edge.From("golfer", Golfer.Type).
			Ref("tournament_odds").
			Unique().
			Required(),
	}
}

func (TournamentOdds) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("tournament", "golfer").
			Fields("vendor").
			Unique(),
	}
}
