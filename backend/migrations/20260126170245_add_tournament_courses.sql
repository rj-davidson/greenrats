-- Modify "placements" table
ALTER TABLE "placements" ALTER COLUMN "created_at" DROP DEFAULT, ALTER COLUMN "updated_at" DROP DEFAULT;
-- Create "tournament_courses" table
CREATE TABLE "tournament_courses" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "rounds" jsonb NULL,
  "course_tournament_courses" uuid NOT NULL,
  "tournament_tournament_courses" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "tournament_courses_courses_tournament_courses" FOREIGN KEY ("course_tournament_courses") REFERENCES "courses" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "tournament_courses_tournaments_tournament_courses" FOREIGN KEY ("tournament_tournament_courses") REFERENCES "tournaments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "tournamentcourse_tournament_tournament_courses_course_tournamen" to table: "tournament_courses"
CREATE UNIQUE INDEX "tournamentcourse_tournament_tournament_courses_course_tournamen" ON "tournament_courses" ("tournament_tournament_courses", "course_tournament_courses");
