-- noinspection SqlNoDataSourceInspectionForFile
CREATE TABLE "users" (
	"user_id" VARCHAR PRIMARY KEY,
	"name" VARCHAR NOT NULL,
	"email" VARCHAR NOT NULL,
	"created_at" TIMESTAMP DEFAULT (now()),
	"role" VARCHAR NOT NULL,
	"cursor" serial NOT NULL
);

CREATE TABLE "project" (
	"project_id" UUID PRIMARY KEY DEFAULT (gen_random_uuid()),
	"project_name" VARCHAR NOT NULL UNIQUE,
	"created_at" TIMESTAMP DEFAULT (now()),
	"cover_image" VARCHAR NOT NULL
);

CREATE TABLE "services" (
	"service_id" UUID PRIMARY KEY DEFAULT (gen_random_uuid()),
	"service_name" VARCHAR NOT NULL,
	"created_at" TIMESTAMP DEFAULT (now())
);

CREATE TABLE "user_to_project" (
	"user_id" VARCHAR,
	"project_id" UUID
);

CREATE TABLE "project_to_service" (
	"project_id" UUID,
	"service_id" UUID
);

CREATE TABLE "client_token" (
	"token" VARCHAR PRIMARY KEY,
	"project_id" UUID
);

ALTER TABLE "user_to_project"
ADD FOREIGN key ("user_id") REFERENCES "users" ("user_id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "user_to_project"
ADD FOREIGN key ("project_id") REFERENCES "project" ("project_id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "project_to_service"
ADD FOREIGN key ("project_id") REFERENCES "project" ("project_id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "project_to_service"
ADD FOREIGN key ("service_id") REFERENCES "services" ("service_id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "client_token"
ADD FOREIGN key ("project_id") REFERENCES "project" ("project_id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "user_to_project"
ADD PRIMARY KEY (
	"user_id",
	"project_id"
);

ALTER TABLE "project_to_service"
ADD PRIMARY KEY (
	"service_id",
	"project_id"
);

/*
service schema
*/
/* news */
CREATE TABLE "news" (
	"news_id" UUID PRIMARY KEY DEFAULT (gen_random_uuid()),
	"project_id" UUID,
	"title" VARCHAR NOT NULL,
	"link" VARCHAR NOT NULL,
	"text" VARCHAR NOT NULL,
	"image" VARCHAR NOT NULL,
	"created_at" TIMESTAMP DEFAULT (now())
);

ALTER TABLE "news"
ADD FOREIGN key ("project_id") REFERENCES "project" ("project_id") ON UPDATE CASCADE;

/* blogs */
CREATE TABLE "blogs" (
	"blog_id" UUID PRIMARY KEY DEFAULT (gen_random_uuid()),
	"user_id" VARCHAR NOT NULL,
	"project_id" UUID NOT NULL,
	"category_id" UUID NOT NULL,
	"author" VARCHAR NOT NULL,
	"title" VARCHAR NOT NULL,
	"cover_image" VARCHAR NOT NULL,
	"short_text" VARCHAR,
	"content" VARCHAR NOT NULL,
	"created_at" TIMESTAMP DEFAULT (now()),
	"updated_at" TIMESTAMP DEFAULT (now())
);

ALTER TABLE "blogs"
ADD FOREIGN key ("project_id") REFERENCES "project" ("project_id") ON UPDATE CASCADE;

ALTER TABLE "blogs"
ADD FOREIGN key ("user_id") REFERENCES "users" ("user_id") ON UPDATE CASCADE;

ALTER TABLE "blogs"
ADD FOREIGN key ("category_id") REFERENCES "category" ("category_id") ON UPDATE CASCADE;

/* category */
CREATE TABLE "category" (
	"category_id" UUID PRIMARY KEY DEFAULT (gen_random_uuid()),
	"parent_id" UUID,
	"project_id" UUID NOT NULL,
	"category_name" VARCHAR NOT NULL,
	"created_at" TIMESTAMP DEFAULT (now())
);

ALTER TABLE "category"
ADD FOREIGN key ("project_id") REFERENCES "project" ("project_id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "category"
ADD FOREIGN key ("parent_id") REFERENCES "category" ("category_id") ON DELETE CASCADE ON UPDATE CASCADE;

/*
custom function and aggregate
*/
CREATE OR REPLACE FUNCTION jsonb_set(
	x JSONB,
	y JSONB,
	p TEXT[],
	e JSONB,
	b BOOLEAN
) returns JSONB language sql AS $$
SELECT CASE WHEN x IS NULL THEN e ELSE jsonb_set(x, p, e, b) END ; $$;

CREATE OR REPLACE AGGREGATE jsonb_set_agg (
	x JSONB,
	p TEXT[],
	e JSONB,
	b BOOLEAN
) (
	stype = JSONB,
	sfunc = jsonb_set
);

/* gallery */
/* album */
CREATE TABLE "album" (
	"album_id" UUID PRIMARY KEY DEFAULT (gen_random_uuid()),
	"project_id" UUID NOT NULL,
	"user_id" VARCHAR NOT NULL,
	"name" VARCHAR NOT NULL,
	"cover" VARCHAR NOT NULL,
	"created_at" TIMESTAMP DEFAULT (now())
)
ALTER TABLE "album"
ADD FOREIGN key ("project_id") REFERENCES "project" ("project_id") ON UPDATE CASCADE;

ALTER TABLE "album"
ADD FOREIGN key ("user_id") REFERENCES "users" ("user_id") ON UPDATE CASCADE;

/* photos */
CREATE TABLE "photos" (
	"photo_id" UUID PRIMARY KEY DEFAULT (gen_random_uuid()),
	"album_id" UUID NOT NULL,
	"path" VARCHAR NOT NULL,
	"created_at" TIMESTAMP DEFAULT (now()),
	"user_id" VARCHAR NOT NULL
)
ALTER TABLE "photos"
ADD FOREIGN key ("album_id") REFERENCES "album" ("album_id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE "photos"
ADD FOREIGN key ("user_id") REFERENCES "users" ("user_id") ON UPDATE CASCADE;

/* documents */
/* cover */
CREATE TABLE "document_cover" (
	"cover_id" UUID PRIMARY KEY DEFAULT (gen_random_uuid()),
	"project_id" UUID NOT NULL,
	"user_id" VARCHAR NOT NULL,
	"name" VARCHAR NOT NULL,
	"created_at" TIMESTAMP DEFAULT (now())
)
ALTER TABLE "document_cover"
ADD FOREIGN key ("project_id") REFERENCES "project" ("project_id") ON UPDATE CASCADE;

ALTER TABLE "document_cover"
ADD FOREIGN key ("user_id") REFERENCES "users" ("user_id") ON UPDATE CASCADE;

/* document files */
CREATE TABLE "documents" (
	"document_id" UUID PRIMARY KEY DEFAULT (gen_random_uuid()),
	"cover_id" UUID NOT NULL,
	"path" VARCHAR NOT NULL,
	"created_at" TIMESTAMP DEFAULT (now()),
	"user_id" VARCHAR NOT NULL,
	"name" VARCHAR NOT NULL
);

ALTER TABLE "documents"
ADD FOREIGN key ("cover_id") REFERENCES "document_cover" ("cover_id") ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE "documents"
ADD FOREIGN key ("user_id") REFERENCES "users" ("user_id") ON UPDATE CASCADE;

/* contact us */
CREATE TABLE "contact_us" (
	id UUID PRIMARY KEY DEFAULT (gen_random_uuid()),
	project_id UUID NOT NULL,
	created_at TIMESTAMP DEFAULT (now()),
	data JSONB NOT NULL
);

ALTER TABLE "contact_us"
ADD FOREIGN key ("project_id") REFERENCES "project" ("project_id") ON UPDATE CASCADE;

/* prise-reports*/
CREATE TABLE "prise-reports" (
	id UUID PRIMARY KEY DEFAULT (gen_random_uuid()),
	project_id UUID NOT NULL,
	created_at TIMESTAMP DEFAULT (now()),
	data JSONB NOT NULL,
	region VARCHAR NOT NULL
);

ALTER TABLE "prise-reports"
ADD FOREIGN key ("project_id") REFERENCES "project" ("project_id") ON UPDATE CASCADE;

