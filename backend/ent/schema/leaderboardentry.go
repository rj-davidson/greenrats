package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type LeaderboardEntry struct {
	ent.Schema
}

func (LeaderboardEntry) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		BaseMixin{},
	}
}

func (LeaderboardEntry) Fields() []ent.Field {
	return []ent.Field{
		field.Int("position").
			Default(0).
			Comment("Leaderboard position (parsed from 'T5' -> 5, 0 = not determined)"),
		field.Bool("cut").
			Default(false).
			Comment("True if golfer missed the cut"),
		field.Int("score").
			Default(0).
			Comment("Score relative to par (negative is under par)"),
		field.Int("total_strokes").
			Default(0).
			Comment("Total strokes taken"),
		field.Int("earnings").
			Default(0).
			Comment("Prize money in dollars"),
		field.Enum("status").
			Values("pending", "active", "withdrawn", "finished").
			Default("pending").
			Comment("Golfer's status in the tournament"),
		field.Int("current_round").
			Default(0).
			Comment("Current round (1-4)"),
		field.Int("thru").
			Default(0).
			Comment("Holes completed in current round"),
	}
}

func (LeaderboardEntry) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tournament", Tournament.Type).
			Ref("leaderboard_entries").
			Unique().
			Required(),
		edge.From("golfer", Golfer.Type).
			Ref("leaderboard_entries").
			Unique().
			Required(),
		edge.To("rounds", Round.Type),
	}
}

func (LeaderboardEntry) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("tournament", "golfer").
			Unique(),
	}
}
