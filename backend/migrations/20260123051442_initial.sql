-- Create "seasons" table
CREATE TABLE "seasons" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "year" bigint NOT NULL,
  "start_date" timestamptz NOT NULL,
  "end_date" timestamptz NOT NULL,
  "is_current" boolean NOT NULL DEFAULT false,
  PRIMARY KEY ("id")
);
-- Create index "seasons_year_key" to table: "seasons"
CREATE UNIQUE INDEX "seasons_year_key" ON "seasons" ("year");
-- Create "sync_status" table
CREATE TABLE "sync_status" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "sync_type" character varying NOT NULL,
  "last_sync_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "sync_status_sync_type_key" to table: "sync_status"
CREATE UNIQUE INDEX "sync_status_sync_type_key" ON "sync_status" ("sync_type");
-- Create "users" table
CREATE TABLE "users" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "workos_id" character varying NOT NULL,
  "email" character varying NOT NULL,
  "display_name" character varying NULL,
  "is_admin" boolean NOT NULL DEFAULT false,
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
  "joining_enabled" boolean NOT NULL DEFAULT true,
  "league_created_by" uuid NOT NULL,
  "season_leagues" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "leagues_seasons_leagues" FOREIGN KEY ("season_leagues") REFERENCES "seasons" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "leagues_users_created_by" FOREIGN KEY ("league_created_by") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "leagues_code_key" to table: "leagues"
CREATE UNIQUE INDEX "leagues_code_key" ON "leagues" ("code");
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
-- Create "courses" table
CREATE TABLE "courses" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "bdl_id" bigint NULL,
  "pga_tour_id" character varying NULL,
  "name" character varying NOT NULL,
  "par" bigint NULL,
  "yardage" bigint NULL,
  "city" character varying NULL,
  "state" character varying NULL,
  "country" character varying NULL,
  PRIMARY KEY ("id")
);
-- Create index "courses_bdl_id_key" to table: "courses"
CREATE UNIQUE INDEX "courses_bdl_id_key" ON "courses" ("bdl_id");
-- Create index "courses_pga_tour_id_key" to table: "courses"
CREATE UNIQUE INDEX "courses_pga_tour_id_key" ON "courses" ("pga_tour_id");
-- Create "course_holes" table
CREATE TABLE "course_holes" (
  "id" uuid NOT NULL,
  "hole_number" bigint NOT NULL,
  "par" bigint NOT NULL,
  "yardage" bigint NULL,
  "course_holes" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "course_holes_courses_holes" FOREIGN KEY ("course_holes") REFERENCES "courses" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "coursehole_hole_number_course_holes" to table: "course_holes"
CREATE UNIQUE INDEX "coursehole_hole_number_course_holes" ON "course_holes" ("hole_number", "course_holes");
-- Create "golfers" table
CREATE TABLE "golfers" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
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
-- Create "tournaments" table
CREATE TABLE "tournaments" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "bdl_id" bigint NULL,
  "pga_tour_id" character varying NULL,
  "name" character varying NOT NULL,
  "start_date" timestamptz NOT NULL,
  "end_date" timestamptz NOT NULL,
  "season_year" bigint NOT NULL,
  "course" character varying NULL,
  "city" character varying NULL,
  "state" character varying NULL,
  "country" character varying NULL,
  "timezone" character varying NULL,
  "pick_window_opens_at" timestamptz NULL,
  "pick_window_closes_at" timestamptz NULL,
  "purse" bigint NULL,
  "course_tournaments" uuid NULL,
  "season_tournaments" uuid NOT NULL,
  "tournament_champion" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "tournaments_courses_tournaments" FOREIGN KEY ("course_tournaments") REFERENCES "courses" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "tournaments_golfers_champion" FOREIGN KEY ("tournament_champion") REFERENCES "golfers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "tournaments_seasons_tournaments" FOREIGN KEY ("season_tournaments") REFERENCES "seasons" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "tournaments_bdl_id_key" to table: "tournaments"
CREATE UNIQUE INDEX "tournaments_bdl_id_key" ON "tournaments" ("bdl_id");
-- Create index "tournaments_pga_tour_id_key" to table: "tournaments"
CREATE UNIQUE INDEX "tournaments_pga_tour_id_key" ON "tournaments" ("pga_tour_id");
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
-- Create "golfer_seasons" table
CREATE TABLE "golfer_seasons" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "scoring_avg" double precision NULL,
  "top_10s" bigint NULL,
  "cuts_made" bigint NULL,
  "events_played" bigint NULL,
  "wins" bigint NULL,
  "earnings" bigint NULL,
  "driving_distance" double precision NULL,
  "driving_accuracy" double precision NULL,
  "gir_pct" double precision NULL,
  "putting_avg" double precision NULL,
  "scrambling_pct" double precision NULL,
  "last_synced_at" timestamptz NULL,
  "golfer_seasons" uuid NOT NULL,
  "season_golfer_seasons" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "golfer_seasons_golfers_seasons" FOREIGN KEY ("golfer_seasons") REFERENCES "golfers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "golfer_seasons_seasons_golfer_seasons" FOREIGN KEY ("season_golfer_seasons") REFERENCES "seasons" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "golferseason_golfer_seasons_season_golfer_seasons" to table: "golfer_seasons"
CREATE UNIQUE INDEX "golferseason_golfer_seasons_season_golfer_seasons" ON "golfer_seasons" ("golfer_seasons", "season_golfer_seasons");
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
  "entry_status" character varying NOT NULL DEFAULT 'confirmed',
  "qualifier" character varying NULL,
  "owgr_at_entry" bigint NULL,
  "is_amateur" boolean NOT NULL DEFAULT false,
  "golfer_entries" uuid NOT NULL,
  "tournament_entries" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "tournament_entries_golfers_entries" FOREIGN KEY ("golfer_entries") REFERENCES "golfers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "tournament_entries_tournaments_entries" FOREIGN KEY ("tournament_entries") REFERENCES "tournaments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "tournamententry_tournament_entries_golfer_entries" to table: "tournament_entries"
CREATE UNIQUE INDEX "tournamententry_tournament_entries_golfer_entries" ON "tournament_entries" ("tournament_entries", "golfer_entries");
-- Create "rounds" table
CREATE TABLE "rounds" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "round_number" bigint NOT NULL,
  "score" bigint NULL,
  "par_relative_score" bigint NULL,
  "tee_time" timestamptz NULL,
  "course_rounds" uuid NULL,
  "tournament_entry_rounds" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "rounds_courses_rounds" FOREIGN KEY ("course_rounds") REFERENCES "courses" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "rounds_tournament_entries_rounds" FOREIGN KEY ("tournament_entry_rounds") REFERENCES "tournament_entries" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "round_round_number_tournament_entry_rounds" to table: "rounds"
CREATE UNIQUE INDEX "round_round_number_tournament_entry_rounds" ON "rounds" ("round_number", "tournament_entry_rounds");
-- Create "hole_scores" table
CREATE TABLE "hole_scores" (
  "id" uuid NOT NULL,
  "hole_number" bigint NOT NULL,
  "par" bigint NOT NULL,
  "score" bigint NULL,
  "round_hole_scores" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "hole_scores_rounds_hole_scores" FOREIGN KEY ("round_hole_scores") REFERENCES "rounds" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "holescore_hole_number_round_hole_scores" to table: "hole_scores"
CREATE UNIQUE INDEX "holescore_hole_number_round_hole_scores" ON "hole_scores" ("hole_number", "round_hole_scores");
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
-- Create "picks" table
CREATE TABLE "picks" (
  "id" uuid NOT NULL,
  "season_year" bigint NOT NULL,
  "created_at" timestamptz NOT NULL,
  "golfer_picks" uuid NOT NULL,
  "league_picks" uuid NOT NULL,
  "season_picks" uuid NOT NULL,
  "tournament_picks" uuid NOT NULL,
  "user_picks" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "picks_golfers_picks" FOREIGN KEY ("golfer_picks") REFERENCES "golfers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "picks_leagues_picks" FOREIGN KEY ("league_picks") REFERENCES "leagues" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "picks_seasons_picks" FOREIGN KEY ("season_picks") REFERENCES "seasons" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "picks_tournaments_picks" FOREIGN KEY ("tournament_picks") REFERENCES "tournaments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "picks_users_picks" FOREIGN KEY ("user_picks") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "pick_season_year_user_picks_golfer_picks_league_picks" to table: "picks"
CREATE UNIQUE INDEX "pick_season_year_user_picks_golfer_picks_league_picks" ON "picks" ("season_year", "user_picks", "golfer_picks", "league_picks");
-- Create index "pick_user_picks_tournament_picks_league_picks" to table: "picks"
CREATE UNIQUE INDEX "pick_user_picks_tournament_picks_league_picks" ON "picks" ("user_picks", "tournament_picks", "league_picks");
