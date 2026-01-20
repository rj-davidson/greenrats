package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type SyncStatus struct {
	ent.Schema
}

func (SyncStatus) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		BaseMixin{},
	}
}

func (SyncStatus) Fields() []ent.Field {
	return []ent.Field{
		field.String("sync_type").NotEmpty().Unique(),
		field.Time("last_sync_at"),
	}
}
