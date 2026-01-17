package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/gofrs/uuid/v5"
)

// GenerateUUID7 generates a new UUID7 (time-sortable, better indexing).
func GenerateUUID7() uuid.UUID {
	id, _ := uuid.NewV7()
	return id
}

// IDMixin provides a UUID7 primary key.
type IDMixin struct {
	mixin.Schema
}

// Fields of the IDMixin.
func (IDMixin) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(GenerateUUID7).
			Immutable(),
	}
}

// BaseMixin provides created_at and updated_at fields.
type BaseMixin struct {
	mixin.Schema
}

// Fields of the BaseMixin.
func (BaseMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// OwnershipMixin provides created_at, updated_at, and created_by edge.
type OwnershipMixin struct {
	mixin.Schema
}

// Fields of the OwnershipMixin.
func (OwnershipMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the OwnershipMixin.
func (OwnershipMixin) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("created_by", User.Type).
			Unique().
			Required(),
	}
}
