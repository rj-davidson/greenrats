-- Step 1: Add new columns to rounds table (nullable initially)
ALTER TABLE "rounds" ADD COLUMN "tournament_rounds" uuid NULL;
ALTER TABLE "rounds" ADD COLUMN "golfer_rounds" uuid NULL;

-- Step 2: Populate new columns from leaderboard_entry relationship
UPDATE "rounds" r
SET
    tournament_rounds = le.tournament_leaderboard_entries,
    golfer_rounds = le.golfer_leaderboard_entries
FROM "leaderboard_entries" le
WHERE r.leaderboard_entry_rounds = le.id;

-- Step 3: Make new columns required and add foreign keys
ALTER TABLE "rounds" ALTER COLUMN "tournament_rounds" SET NOT NULL;
ALTER TABLE "rounds" ALTER COLUMN "golfer_rounds" SET NOT NULL;

ALTER TABLE "rounds"
    ADD CONSTRAINT "rounds_tournaments_rounds" FOREIGN KEY ("tournament_rounds") REFERENCES "tournaments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
    ADD CONSTRAINT "rounds_golfers_rounds" FOREIGN KEY ("golfer_rounds") REFERENCES "golfers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

-- Step 4: Drop old index and constraint from rounds
DROP INDEX IF EXISTS "round_round_number_leaderboard_entry_rounds";
ALTER TABLE "rounds" DROP CONSTRAINT IF EXISTS "rounds_leaderboard_entries_rounds";

-- Step 5: Drop old column from rounds
ALTER TABLE "rounds" DROP COLUMN "leaderboard_entry_rounds";

-- Step 6: Create new unique index on rounds
CREATE UNIQUE INDEX "round_round_number_tournament_rounds_golfer_rounds" ON "rounds" ("round_number", "tournament_rounds", "golfer_rounds");

-- Step 7: Create placements table
CREATE TABLE "placements" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "position" character varying NOT NULL DEFAULT '',
  "position_numeric" bigint NULL,
  "total_score" bigint NULL,
  "par_relative_score" bigint NULL,
  "earnings" bigint NOT NULL DEFAULT 0,
  "status" character varying NOT NULL DEFAULT 'finished',
  "golfer_placements" uuid NOT NULL,
  "tournament_placements" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "placements_golfers_placements" FOREIGN KEY ("golfer_placements") REFERENCES "golfers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "placements_tournaments_placements" FOREIGN KEY ("tournament_placements") REFERENCES "tournaments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);

-- Step 8: Create unique index on placements
CREATE UNIQUE INDEX "placement_tournament_placements_golfer_placements" ON "placements" ("tournament_placements", "golfer_placements");

-- Step 9: Migrate data from leaderboard_entries to placements
INSERT INTO "placements" (
    id,
    created_at,
    updated_at,
    position,
    position_numeric,
    total_score,
    par_relative_score,
    earnings,
    status,
    golfer_placements,
    tournament_placements
)
SELECT
    id,
    created_at,
    updated_at,
    CASE
        WHEN cut = true THEN 'CUT'
        WHEN position > 0 THEN position::varchar
        ELSE ''
    END,
    CASE WHEN cut = true THEN NULL ELSE position END,
    total_strokes,
    score,
    earnings,
    CASE
        WHEN cut = true THEN 'cut'
        WHEN status = 'withdrawn' THEN 'withdrawn'
        ELSE 'finished'
    END,
    golfer_leaderboard_entries,
    tournament_leaderboard_entries
FROM "leaderboard_entries";

-- Step 10: Drop leaderboard_entries table
DROP TABLE "leaderboard_entries";
