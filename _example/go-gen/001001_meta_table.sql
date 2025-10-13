-- SEQUENCE
CREATE SEQUENCE IF NOT EXISTS "meta__id__seq" INCREMENT 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1;

-- TABLE
CREATE TABLE IF NOT EXISTS "meta"
(
	"id" BIGINT DEFAULT nextval('meta__id__seq') NOT NULL,
	CONSTRAINT "meta__id__pk" PRIMARY KEY ( "id" ),
	"uid" UUID NOT NULL,
	"user_id" BIGINT NOT NULL,
	"roles" TEXT[] NOT NULL,
	CONSTRAINT "meta__roles__unq" UNIQUE ( "roles" ),
	"fail" BOOLEAN NOT NULL,
	"created_at" TIMESTAMPTZ NOT NULL,
	"updated_at" TIMESTAMPTZ NOT NULL,
	"deleted_at" TIMESTAMPTZ NULL
);

-- INDEX
CREATE INDEX "meta__uid__idx" ON "meta" USING btree ( "uid" );
ALTER TABLE "meta" ADD CONSTRAINT "meta__id_user_id__unq" UNIQUE ("id", "user_id");
ALTER TABLE "meta" ADD CONSTRAINT "meta__user_id_id__unq" UNIQUE ("user_id", "id");

