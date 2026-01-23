package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type FieldEntry struct {
	ent.Schema
}

func (FieldEntry) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		BaseMixin{},
	}
}

func (FieldEntry) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("entry_status").
			Values("confirmed", "alternate", "withdrawn", "pending").
			Default("confirmed").
			Comment("Field entry status"),
		field.String("qualifier").
			Optional().
			Nillable().
			Comment("Qualification category (e.g., 'winner', 'exemption', 'sponsor')"),
		field.Int("owgr_at_entry").
			Optional().
			Nillable().
			Comment("Official World Golf Ranking at time of field entry"),
		field.Bool("is_amateur").
			Default(false).
			Comment("True if golfer is playing as an amateur"),
	}
}

func (FieldEntry) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tournament", Tournament.Type).
			Ref("field_entries").
			Unique().
			Required(),
		edge.From("golfer", Golfer.Type).
			Ref("field_entries").
			Unique().
			Required(),
	}
}

func (FieldEntry) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("tournament", "golfer").
			Unique(),
	}
}
