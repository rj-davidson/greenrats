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
