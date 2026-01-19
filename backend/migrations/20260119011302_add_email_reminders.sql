-- Create "email_reminders" table
CREATE TABLE "email_reminders" (
  "id" uuid NOT NULL,
  "reminder_type" character varying NOT NULL,
  "sent_at" timestamptz NOT NULL,
  "league_email_reminders" uuid NOT NULL,
  "tournament_email_reminders" uuid NOT NULL,
  "user_email_reminders" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "email_reminders_leagues_email_reminders" FOREIGN KEY ("league_email_reminders") REFERENCES "leagues" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "email_reminders_tournaments_email_reminders" FOREIGN KEY ("tournament_email_reminders") REFERENCES "tournaments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "email_reminders_users_email_reminders" FOREIGN KEY ("user_email_reminders") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "emailreminder_reminder_type_user_email_reminders_tournament_ema" to table: "email_reminders"
CREATE UNIQUE INDEX "emailreminder_reminder_type_user_email_reminders_tournament_ema" ON "email_reminders" ("reminder_type", "user_email_reminders", "tournament_email_reminders", "league_email_reminders");
