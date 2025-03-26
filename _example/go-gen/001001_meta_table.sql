-- TABLE
CREATE TABLE IF NOT EXISTS "meta" (
	 "id" UUID NOT NULL ,
	 CONSTRAINT "meta_id_pk" PRIMARY KEY ("id"),
	 "user_id" BIGINT NOT NULL ,
	 CONSTRAINT "meta_user_id_fk" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE NOT DEFERRABLE,
	 "roles" TEXT[] NOT NULL ,
	 CONSTRAINT "meta_roles_unq" UNIQUE ("roles"),
	 "fail" BOOLEAN NOT NULL,
	 "created_at" TIMESTAMPTZ NOT NULL,
	 "updated_at" TIMESTAMPTZ NOT NULL,
	 "deleted_at" TIMESTAMPTZ NULL,
	 CONSTRAINT "meta__id_user_id__uniq" UNIQUE ("id","user_id"),
	 CONSTRAINT "meta__user_id_id__uniq" UNIQUE ("user_id","id")
);

