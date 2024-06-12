CREATE TABLE "users" (
  "user_id" varchar PRIMARY KEY,
  "name" varchar NOT NULL,
  "email" varchar NOT NULL,
  "created_at" timestamp DEFAULT (now()),
  "role" varchar NOT NULL,
  "cursor" serial NOT NULL
);

CREATE TABLE "project" (
  "project_id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "project_name" varchar NOT NULL UNIQUE,
  "created_at" timestamp DEFAULT (now())
);

CREATE TABLE "services" (
  "service_id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "service_name" varchar NOT NULL,
  "created_at" timestamp DEFAULT (now())
);

CREATE TABLE "user_to_project" (
  "user_id" varchar,
  "project_id" uuid
);

CREATE TABLE "project_to_service" (
  "project_id" uuid,
  "service_id" uuid
);

CREATE TABLE "client_token" (
  "token" varchar PRIMARY KEY,
  "project_id" uuid
);

ALTER TABLE "user_to_project" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("user_id") on delete cascade on update cascade;

ALTER TABLE "user_to_project" ADD FOREIGN KEY ("project_id") REFERENCES "project" ("project_id") on delete cascade on update cascade;

ALTER TABLE "project_to_service" ADD FOREIGN KEY ("project_id") REFERENCES "project" ("project_id") on delete cascade on update cascade;

ALTER TABLE "project_to_service" ADD FOREIGN KEY ("service_id") REFERENCES "services" ("service_id") on delete cascade on update cascade;

ALTER TABLE "client_token" ADD FOREIGN KEY ("project_id") REFERENCES "project" ("project_id") on delete cascade on update cascade;


/* 
    service schema
*/

/* news */
CREATE TABLE "news" (
    "news_id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
    "project_id" uuid,
    "title" varchar NOT NULL,
    "link" varchar NOT NULL,
    "text" varchar NOT NULL,
    "image" varchar NOT NULL,
    "created_at" timestamp DEFAULT (now())
);

ALTER TABLE "news" ADD FOREIGN KEY ("project_id") REFERENCES "project" ("project_id") on delete cascade on update cascade;