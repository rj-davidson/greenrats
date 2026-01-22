package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// GolferSeason holds the schema definition for the GolferSeason entity.
type GolferSeason struct {
	ent.Schema
}

// Mixin of the GolferSeason.
func (GolferSeason) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		BaseMixin{},
	}
}

// Fields of the GolferSeason.
func (GolferSeason) Fields() []ent.Field {
	return []ent.Field{
		field.Float("scoring_avg").
			Optional().
			Nillable(),
		field.Int("top_10s").
			Optional().
			Nillable(),
		field.Int("cuts_made").
			Optional().
			Nillable(),
		field.Int("events_played").
			Optional().
			Nillable(),
		field.Int("wins").
			Optional().
			Nillable(),
		field.Int("earnings").
			Optional().
			Nillable().
			Comment("Total prize money in dollars"),
		field.Float("driving_distance").
			Optional().
			Nillable(),
		field.Float("driving_accuracy").
			Optional().
			Nillable().
			Comment("Percentage"),
		field.Float("gir_pct").
			Optional().
			Nillable().
			Comment("Greens in regulation percentage"),
		field.Float("putting_avg").
			Optional().
			Nillable(),
		field.Float("scrambling_pct").
			Optional().
			Nillable().
			Comment("Scrambling percentage"),
		field.Time("last_synced_at").
			Optional().
			Nillable(),
	}
}

// Edges of the GolferSeason.
func (GolferSeason) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("golfer", Golfer.Type).
			Ref("seasons").
			Unique().
			Required(),
		edge.From("season", Season.Type).
			Ref("golfer_seasons").
			Unique().
			Required(),
	}
}

// Indexes of the GolferSeason.
func (GolferSeason) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("golfer", "season").
			Unique(),
	}
}
