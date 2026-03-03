package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Golfer struct {
	ent.Schema
}

func (Golfer) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		BaseMixin{},
	}
}

func (Golfer) Fields() []ent.Field {
	return []ent.Field{
		field.Int("bdl_id").
			Optional().
			Nillable().
			Unique().
			Comment("BallDontLie API ID (int)"),
		field.String("first_name").
			Optional().
			Nillable(),
		field.String("last_name").
			Optional().
			Nillable(),
		field.String("name").
			NotEmpty().
			Comment("Display name"),
		field.String("country").
			Optional().
			Nillable().
			Comment("Full country name"),
		field.String("country_code").
			Default("UNK").
			Comment("3-letter ISO country code"),
		field.Int("owgr").
			Optional().
			Nillable().
			Comment("Official World Golf Ranking"),
		field.Bool("active").
			Default(true).
			Comment("Whether the player is currently active"),
		field.String("image_url").
			Optional().
			Nillable(),
		field.String("height").
			Optional().
			Nillable().
			Comment("Height (e.g., '6-1')"),
		field.String("weight").
			Optional().
			Nillable().
			Comment("Weight (e.g., '175 lbs')"),
		field.Time("birth_date").
			Optional().
			Nillable(),
		field.String("birthplace_city").
			Optional().
			Nillable(),
		field.String("birthplace_state").
			Optional().
			Nillable(),
		field.String("birthplace_country").
			Optional().
			Nillable(),
		field.Int("turned_pro").
			Optional().
			Nillable().
			Comment("Year turned professional"),
		field.String("school").
			Optional().
			Nillable().
			Comment("College/university attended"),
		field.String("residence_city").
			Optional().
			Nillable(),
		field.String("residence_state").
			Optional().
			Nillable(),
		field.String("residence_country").
			Optional().
			Nillable(),
	}
}

func (Golfer) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("picks", Pick.Type),
		edge.To("field_entries", FieldEntry.Type),
		edge.To("placements", Placement.Type),
		edge.To("rounds", Round.Type),
		edge.To("seasons", GolferSeason.Type),
		edge.To("tournament_odds", TournamentOdds.Type),
	}
}
