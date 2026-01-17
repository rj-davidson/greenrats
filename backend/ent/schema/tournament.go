package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Tournament holds the schema definition for the Tournament entity.
type Tournament struct {
	ent.Schema
}

// Mixin of the Tournament.
func (Tournament) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

// Fields of the Tournament.
func (Tournament) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("external_id").
			Unique().
			NotEmpty(),
		field.String("name").
			NotEmpty(),
		field.Time("start_date"),
		field.Time("end_date"),
		field.Enum("status").
			Values("upcoming", "active", "completed").
			Default("upcoming"),
		field.Int("season_year"),
	}
}

// Edges of the Tournament.
func (Tournament) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("picks", Pick.Type),
		edge.To("golfers", Golfer.Type),
	}
}
