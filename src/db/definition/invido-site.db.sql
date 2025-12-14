BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS "post" (
	"id"	INTEGER,
	"title"	TEXT,
	"post_id"	TEXT NOT NULL,
	"timestamp"	NUMERIC,
	"abstract"	TEXT,
	"uri"	TEXT,
	-- "next_post_id"	TEXT,
	-- "prev_post_id"	TEXT,
	-- "content"	TEXT,
	-- "status" INTEGER,
	-- "tags"	TEXT,
	PRIMARY KEY("id" AUTOINCREMENT)
);
COMMIT;
