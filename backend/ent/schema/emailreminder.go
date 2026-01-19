package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type EmailReminder struct {
	ent.Schema
}

func (EmailReminder) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
	}
}

func (EmailReminder) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("reminder_type").
			Values("pick_reminder", "tournament_results").
			Immutable(),
		field.Time("sent_at").
			Default(time.Now).
			Immutable(),
	}
}

func (EmailReminder) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("email_reminders").
			Unique().
			Required(),
		edge.From("tournament", Tournament.Type).
			Ref("email_reminders").
			Unique().
			Required(),
		edge.From("league", League.Type).
			Ref("email_reminders").
			Unique().
			Required(),
	}
}

func (EmailReminder) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("user", "tournament", "league").
			Fields("reminder_type").
			Unique(),
	}
}
