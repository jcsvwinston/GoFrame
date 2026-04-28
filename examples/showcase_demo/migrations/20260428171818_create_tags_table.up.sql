CREATE TABLE IF NOT EXISTS "tags" (
	"id" INTEGER PRIMARY KEY AUTOINCREMENT,
	"created_at" DATETIME,
	"updated_at" DATETIME,
	"deleted_at" DATETIME,
	"name" TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS "idx_tags_name" ON "tags" ("name");
