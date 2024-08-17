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
  "created_at" timestamp DEFAULT (now()),
  "cover_image" varchar NOT Null
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

ALTER TABLE "user_to_project" ADD PRIMARY KEY ("user_id", "project_id");

ALTER TABLE "project_to_service" ADD PRIMARY KEY ("service_id", "project_id");


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

ALTER TABLE "news" ADD FOREIGN KEY ("project_id") REFERENCES "project" ("project_id") on update cascade;

/* blogs */
CREATE TABLE "blogs" (
  "blog_id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "user_id" varchar NOT NULL,
  "project_id" uuid NOT NULL,
  "category_id" uuid NOT NULL,
  "author" varchar NOT NULL,
  "title" varchar NOT NULL,
  "cover_image" varchar NOT NULL,
  "short_text" varchar,
  "content" varchar NOT NULL,
  "created_at" timestamp DEFAULT (now()),
  "updated_at" timestamp DEFAULT (now())
);

ALTER TABLE "blogs" ADD FOREIGN KEY ("project_id") REFERENCES "project" ("project_id") on update cascade;
ALTER TABLE "blogs" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("user_id") on update cascade;
ALTER TABLE "blogs" ADD FOREIGN KEY ("category_id") REFERENCES "category" ("category_id") on update cascade;

/* category */
CREATE TABLE "category" (
  "category_id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "parent_id" uuid,
  "project_id" uuid NOT NULL,
  "category_name" varchar NOT NULL,
  "created_at" timestamp DEFAULT (now())
);

ALTER TABLE "category" ADD FOREIGN KEY ("project_id") REFERENCES "project" ("project_id") on update cascade;
ALTER TABLE "category" ADD FOREIGN KEY ("parent_id") REFERENCES "category" ("category_id") on delete cascade on update cascade;


/* 
    custom function and aggregate
*/
CREATE OR REPLACE FUNCTION jsonb_set(x jsonb, y jsonb, p text[], e jsonb, b boolean)
RETURNS jsonb LANGUAGE sql AS $$
SELECT CASE WHEN x IS NULL THEN e ELSE jsonb_set(x, p, e, b) END ; $$ ;

CREATE OR REPLACE AGGREGATE jsonb_set_agg(x jsonb, p text[], e jsonb, b boolean)
( STYPE = jsonb, SFUNC = jsonb_set);


/* gallery */

/* album */
CREATE TABLE "album" (
	"album_id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
	"project_id" uuid NOT NULL,
    "user_id" varchar NOT NULL,
	"name" varchar NOT NULL,
	"cover" varchar NOT NULL,
	"created_at" timestamp DEFAULT(now())
)

ALTER TABLE "album" ADD FOREIGN KEY ("project_id") REFERENCES "project" ("project_id") on update cascade;
ALTER TABLE "album" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("user_id") on update cascade;