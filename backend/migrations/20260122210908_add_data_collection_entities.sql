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
-- Create "rounds" table
CREATE TABLE "rounds" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "round_number" bigint NOT NULL,
  "score" bigint NULL,
  "par_relative_score" bigint NULL,
  "tee_time" timestamptz NULL,
  "tournament_entry_rounds" uuid NOT NULL,
  PRIMARY KEY ("id"),
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
-- Modify "leagues" table
ALTER TABLE "leagues" ADD COLUMN "season_leagues" uuid NULL, ADD CONSTRAINT "leagues_seasons_leagues" FOREIGN KEY ("season_leagues") REFERENCES "seasons" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "picks" table
ALTER TABLE "picks" ADD COLUMN "season_picks" uuid NULL, ADD CONSTRAINT "picks_seasons_picks" FOREIGN KEY ("season_picks") REFERENCES "seasons" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "tournaments" table
ALTER TABLE "tournaments" ADD COLUMN "course_tournaments" uuid NULL, ADD COLUMN "season_tournaments" uuid NULL, ADD CONSTRAINT "tournaments_courses_tournaments" FOREIGN KEY ("course_tournaments") REFERENCES "courses" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tournaments_seasons_tournaments" FOREIGN KEY ("season_tournaments") REFERENCES "seasons" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
