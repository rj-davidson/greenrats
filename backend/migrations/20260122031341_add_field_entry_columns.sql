-- Modify "tournament_entries" table
ALTER TABLE "tournament_entries" ADD COLUMN "entry_status" character varying NOT NULL DEFAULT 'confirmed', ADD COLUMN "qualifier" character varying NULL, ADD COLUMN "owgr_at_entry" bigint NULL, ADD COLUMN "is_amateur" boolean NOT NULL DEFAULT false;
