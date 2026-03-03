-- Create "tournament_odds" table
CREATE TABLE "tournament_odds" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "vendor" character varying NOT NULL,
  "american_odds" bigint NOT NULL,
  "implied_probability" double precision NOT NULL,
  "odds_updated_at" timestamptz NOT NULL,
  "golfer_tournament_odds" uuid NOT NULL,
  "tournament_tournament_odds" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "tournament_odds_golfers_tournament_odds" FOREIGN KEY ("golfer_tournament_odds") REFERENCES "golfers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "tournament_odds_tournaments_tournament_odds" FOREIGN KEY ("tournament_tournament_odds") REFERENCES "tournaments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "tournamentodds_vendor_tournament_tournament_odds_golfer_tournam" to table: "tournament_odds"
CREATE UNIQUE INDEX "tournamentodds_vendor_tournament_tournament_odds_golfer_tournam" ON "tournament_odds" ("vendor", "tournament_tournament_odds", "golfer_tournament_odds");
