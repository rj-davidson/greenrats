-- Modify "tournaments" table
ALTER TABLE "tournaments" ADD COLUMN "pga_tour_id" character varying NULL;
-- Create index "tournaments_pga_tour_id_key" to table: "tournaments"
CREATE UNIQUE INDEX "tournaments_pga_tour_id_key" ON "tournaments" ("pga_tour_id");
