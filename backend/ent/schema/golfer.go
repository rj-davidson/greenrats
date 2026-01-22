package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Golfer holds the schema definition for the Golfer entity.
type Golfer struct {
	ent.Schema
}

// Mixin of the Golfer.
func (Golfer) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		BaseMixin{},
	}
}

// Fields of the Golfer.
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
	}
}

// Edges of the Golfer.
func (Golfer) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("picks", Pick.Type),
		edge.To("entries", TournamentEntry.Type),
	}
}
