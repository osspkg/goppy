-- SEQUENCE
CREATE SEQUENCE IF NOT EXISTS "users_id_seq" INCREMENT 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1;

-- TABLE
CREATE TABLE IF NOT EXISTS "users" (
	 "id" BIGINT DEFAULT nextval('users_id_seq') NOT NULL NOT NULL ,
	 CONSTRAINT "users_id_pk" PRIMARY KEY ("id"),
	 "name" VARCHAR(100) NOT NULL,
	 "value" TEXT[] NOT NULL
);

