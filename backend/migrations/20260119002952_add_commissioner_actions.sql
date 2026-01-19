-- Create "commissioner_actions" table
CREATE TABLE "commissioner_actions" (
  "id" uuid NOT NULL,
  "action_type" character varying NOT NULL,
  "description" character varying NOT NULL,
  "metadata" jsonb NULL,
  "created_at" timestamptz NOT NULL,
  "league_commissioner_actions" uuid NOT NULL,
  "user_commissioner_actions" uuid NOT NULL,
  "user_affected_actions" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "commissioner_actions_leagues_commissioner_actions" FOREIGN KEY ("league_commissioner_actions") REFERENCES "leagues" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "commissioner_actions_users_affected_actions" FOREIGN KEY ("user_affected_actions") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "commissioner_actions_users_commissioner_actions" FOREIGN KEY ("user_commissioner_actions") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
