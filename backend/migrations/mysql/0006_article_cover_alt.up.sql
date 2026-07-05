-- Alt text for the cover image, for accessibility (screen readers) and SEO.
-- Nullable: existing articles and cover-less drafts have none. Bounded to 300
-- chars to match the in-article image alt limit.

ALTER TABLE articles ADD COLUMN cover_alt VARCHAR(300) AFTER cover;
