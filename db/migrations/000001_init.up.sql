CREATE TABLE "users" (
  "id" int PRIMARY KEY,
  "name" varchar UNIQUE NOT NULL,
  "passhash" varchar NOT NULL
);

CREATE TABLE "secrets" (
  "id" bigint PRIMARY KEY,
  "owner" varchar,
  "kind" int,
  "name" varchar,
  "value" bytea,
  "created" timestamptz DEFAULT (now()),
  "modified" timestamptz DEFAULT (now()),
  "deleted" boolean DEFAULT false
);

ALTER TABLE "secrets" ADD FOREIGN KEY ("owner") REFERENCES "users" ("name");
