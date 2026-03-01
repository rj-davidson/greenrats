package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Tournament struct {
	ent.Schema
}

func (Tournament) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		BaseMixin{},
	}
}

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
		field.Int("season_year"),
		field.String("course").
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
		field.String("purse").
			Optional().
			Nillable().
			Comment("Total prize money as formatted string from API"),
	}
}

func (Tournament) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("picks", Pick.Type),
		edge.To("field_entries", FieldEntry.Type),
		edge.To("placements", Placement.Type),
		edge.To("rounds", Round.Type),
		edge.To("email_reminders", EmailReminder.Type),
		edge.To("tournament_courses", TournamentCourse.Type),
		edge.To("champion", Golfer.Type).
			Unique(),
		edge.From("season", Season.Type).
			Ref("tournaments").
			Unique().
			Required(),
		edge.From("course_ref", Course.Type).
			Ref("tournaments").
			Unique(),
	}
}
