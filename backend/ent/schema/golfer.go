package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Golfer holds the schema definition for the Golfer entity.
type Golfer struct {
	ent.Schema
}

// Mixin of the Golfer.
func (Golfer) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

// Fields of the Golfer.
func (Golfer) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("external_id").
			Unique().
			NotEmpty(),
		field.String("name").
			NotEmpty(),
		field.String("country").
			NotEmpty(),
		field.Int("world_ranking").
			Optional().
			Nillable(),
		field.String("image_url").
			Optional().
			Nillable(),
	}
}

// Edges of the Golfer.
func (Golfer) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("picks", Pick.Type),
		edge.From("tournaments", Tournament.Type).
			Ref("golfers"),
	}
}
