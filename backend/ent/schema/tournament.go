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
		field.String("scratchgolf_id").
			Optional().
			Nillable().
			Unique().
			Comment("ScratchGolf API ID (string)"),
		field.Int("bdl_id").
			Optional().
			Nillable().
			Unique().
			Comment("BallDontLie API ID (int)"),
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
			Nillable(),
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
	}
}
