package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Tournament holds the schema definition for the Tournament entity.
type Tournament struct {
	ent.Schema
}

// Mixin of the Tournament.
func (Tournament) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		BaseMixin{},
	}
}

// Fields of the Tournament.
func (Tournament) Fields() []ent.Field {
	return []ent.Field{
		field.Int("bdl_id").
			Optional().
			Nillable().
			Unique().
			Comment("BallDontLie API ID (int)"),
		field.String("pga_tour_id").
			Optional().
			Nillable().
			Unique().
			Comment("PGA Tour tournament ID (e.g., R2025002)"),
		field.String("name").
			NotEmpty(),
		field.Time("start_date"),
		field.Time("end_date"),
		field.Enum("status").
			Values("upcoming", "active", "completed").
			Default("upcoming"),
		field.Int("season_year"),
		field.String("course").
			Optional().
			Nillable(),
		field.String("location").
			Optional().
			Nillable().
			Comment("Deprecated: use city/state/country instead"),
		field.String("city").
			Optional().
			Nillable(),
		field.String("state").
			Optional().
			Nillable(),
		field.String("country").
			Optional().
			Nillable(),
		field.String("timezone").
			Optional().
			Nillable().
			Comment("IANA timezone (e.g., America/New_York)"),
		field.Time("pick_window_opens_at").
			Optional().
			Nillable().
			Comment("UTC timestamp when pick window opens"),
		field.Time("pick_window_closes_at").
			Optional().
			Nillable().
			Comment("UTC timestamp when pick window closes"),
		field.Int("purse").
			Optional().
			Nillable().
			Comment("Total prize money in dollars"),
	}
}

// Edges of the Tournament.
func (Tournament) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("picks", Pick.Type),
		edge.To("entries", TournamentEntry.Type),
		edge.To("email_reminders", EmailReminder.Type),
		edge.To("champion", Golfer.Type).
			Unique(),
	}
}
