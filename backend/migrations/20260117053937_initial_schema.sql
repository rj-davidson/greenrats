-- Create "users" table
CREATE TABLE "users" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "workos_id" character varying NOT NULL,
  "email" character varying NOT NULL,
  "display_name" character varying NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "users_workos_id_key" to table: "users"
CREATE UNIQUE INDEX "users_workos_id_key" ON "users" ("workos_id");
-- Create "leagues" table
CREATE TABLE "leagues" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "code" character varying NOT NULL,
  "season_year" bigint NOT NULL,
  "league_created_by" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "leagues_users_created_by" FOREIGN KEY ("league_created_by") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "leagues_code_key" to table: "leagues"
CREATE UNIQUE INDEX "leagues_code_key" ON "leagues" ("code");
-- Create "league_memberships" table
CREATE TABLE "league_memberships" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "role" character varying NOT NULL DEFAULT 'member',
  "joined_at" timestamptz NOT NULL,
  "league_memberships" uuid NOT NULL,
  "league_membership_created_by" uuid NOT NULL,
  "user_league_memberships" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "league_memberships_leagues_memberships" FOREIGN KEY ("league_memberships") REFERENCES "leagues" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "league_memberships_users_created_by" FOREIGN KEY ("league_membership_created_by") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "league_memberships_users_league_memberships" FOREIGN KEY ("user_league_memberships") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "leaguemembership_user_league_memberships_league_memberships" to table: "league_memberships"
CREATE UNIQUE INDEX "leaguemembership_user_league_memberships_league_memberships" ON "league_memberships" ("user_league_memberships", "league_memberships");
-- Create "golfers" table
CREATE TABLE "golfers" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "external_id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "country" character varying NOT NULL,
  "world_ranking" bigint NULL,
  "image_url" character varying NULL,
  PRIMARY KEY ("id")
);
-- Create index "golfers_external_id_key" to table: "golfers"
CREATE UNIQUE INDEX "golfers_external_id_key" ON "golfers" ("external_id");
-- Create "tournaments" table
CREATE TABLE "tournaments" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "external_id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "start_date" timestamptz NOT NULL,
  "end_date" timestamptz NOT NULL,
  "status" character varying NOT NULL DEFAULT 'upcoming',
  "season_year" bigint NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "tournaments_external_id_key" to table: "tournaments"
CREATE UNIQUE INDEX "tournaments_external_id_key" ON "tournaments" ("external_id");
-- Create "picks" table
CREATE TABLE "picks" (
  "id" uuid NOT NULL,
  "season_year" bigint NOT NULL,
  "created_at" timestamptz NOT NULL,
  "golfer_picks" uuid NOT NULL,
  "league_picks" uuid NOT NULL,
  "tournament_picks" uuid NOT NULL,
  "user_picks" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "picks_golfers_picks" FOREIGN KEY ("golfer_picks") REFERENCES "golfers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "picks_leagues_picks" FOREIGN KEY ("league_picks") REFERENCES "leagues" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "picks_tournaments_picks" FOREIGN KEY ("tournament_picks") REFERENCES "tournaments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "picks_users_picks" FOREIGN KEY ("user_picks") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "pick_season_year_user_picks_golfer_picks_league_picks" to table: "picks"
CREATE UNIQUE INDEX "pick_season_year_user_picks_golfer_picks_league_picks" ON "picks" ("season_year", "user_picks", "golfer_picks", "league_picks");
-- Create index "pick_user_picks_tournament_picks_league_picks" to table: "picks"
CREATE UNIQUE INDEX "pick_user_picks_tournament_picks_league_picks" ON "picks" ("user_picks", "tournament_picks", "league_picks");
-- Create "tournament_golfers" table
CREATE TABLE "tournament_golfers" (
  "tournament_id" uuid NOT NULL,
  "golfer_id" uuid NOT NULL,
  PRIMARY KEY ("tournament_id", "golfer_id"),
  CONSTRAINT "tournament_golfers_golfer_id" FOREIGN KEY ("golfer_id") REFERENCES "golfers" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "tournament_golfers_tournament_id" FOREIGN KEY ("tournament_id") REFERENCES "tournaments" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
