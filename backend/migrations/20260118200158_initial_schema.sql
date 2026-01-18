-- Create "users" table
CREATE TABLE "users" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "workos_id" character varying NOT NULL,
  "email" character varying NOT NULL,
  "display_name" character varying NULL,
  PRIMARY KEY ("id")
);
-- Create index "users_display_name_key" to table: "users"
CREATE UNIQUE INDEX "users_display_name_key" ON "users" ("display_name");
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
  "user_league_memberships" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "league_memberships_leagues_memberships" FOREIGN KEY ("league_memberships") REFERENCES "leagues" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "league_memberships_users_league_memberships" FOREIGN KEY ("user_league_memberships") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "leaguemembership_user_league_memberships_league_memberships" to table: "league_memberships"
CREATE UNIQUE INDEX "leaguemembership_user_league_memberships_league_memberships" ON "league_memberships" ("user_league_memberships", "league_memberships");
-- Create "golfers" table
CREATE TABLE "golfers" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "scratchgolf_id" character varying NULL,
  "bdl_id" bigint NULL,
  "first_name" character varying NULL,
  "last_name" character varying NULL,
  "name" character varying NOT NULL,
  "country" character varying NULL,
  "country_code" character varying NOT NULL DEFAULT 'UNK',
  "owgr" bigint NULL,
  "active" boolean NOT NULL DEFAULT true,
  "image_url" character varying NULL,
  PRIMARY KEY ("id")
);
-- Create index "golfers_bdl_id_key" to table: "golfers"
CREATE UNIQUE INDEX "golfers_bdl_id_key" ON "golfers" ("bdl_id");
-- Create index "golfers_scratchgolf_id_key" to table: "golfers"
CREATE UNIQUE INDEX "golfers_scratchgolf_id_key" ON "golfers" ("scratchgolf_id");
-- Create "tournaments" table
CREATE TABLE "tournaments" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "scratchgolf_id" character varying NULL,
  "bdl_id" bigint NULL,
  "name" character varying NOT NULL,
  "start_date" timestamptz NOT NULL,
  "end_date" timestamptz NOT NULL,
  "status" character varying NOT NULL DEFAULT 'upcoming',
  "season_year" bigint NOT NULL,
  "course" character varying NULL,
  "location" character varying NULL,
  "purse" bigint NULL,
  PRIMARY KEY ("id")
);
-- Create index "tournaments_bdl_id_key" to table: "tournaments"
CREATE UNIQUE INDEX "tournaments_bdl_id_key" ON "tournaments" ("bdl_id");
-- Create index "tournaments_scratchgolf_id_key" to table: "tournaments"
CREATE UNIQUE INDEX "tournaments_scratchgolf_id_key" ON "tournaments" ("scratchgolf_id");
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
-- Create "tournament_entries" table
CREATE TABLE "tournament_entries" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "position" bigint NOT NULL DEFAULT 0,
  "cut" boolean NOT NULL DEFAULT false,
  "score" bigint NOT NULL DEFAULT 0,
  "total_strokes" bigint NOT NULL DEFAULT 0,
  "earnings" bigint NOT NULL DEFAULT 0,
  "status" character varying NOT NULL DEFAULT 'pending',
  "current_round" bigint NOT NULL DEFAULT 0,
  "thru" bigint NOT NULL DEFAULT 0,
  "golfer_entries" uuid NOT NULL,
  "tournament_entries" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "tournament_entries_golfers_entries" FOREIGN KEY ("golfer_entries") REFERENCES "golfers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "tournament_entries_tournaments_entries" FOREIGN KEY ("tournament_entries") REFERENCES "tournaments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "tournamententry_tournament_entries_golfer_entries" to table: "tournament_entries"
CREATE UNIQUE INDEX "tournamententry_tournament_entries_golfer_entries" ON "tournament_entries" ("tournament_entries", "golfer_entries");
