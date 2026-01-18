-- Drop existing case-sensitive unique index
DROP INDEX IF EXISTS "users_display_name_key";

-- Create case-insensitive unique index
CREATE UNIQUE INDEX "users_display_name_lower_key" ON "users" (LOWER("display_name"));
