-- Bilingual articles. `lang` names the language of the base content columns
-- (title, excerpt, body, cover_alt); `translations` holds the OTHER language
-- variants keyed by locale, e.g. {"en": {"title","excerpt","coverAlt","body"}}.
-- The cover image, category, tags and series placement stay shared across
-- languages (only their alt text and the prose are translated).
-- Existing rows are Vietnamese, so lang defaults to 'vi' and translations is NULL.

ALTER TABLE articles ADD COLUMN lang VARCHAR(5) NOT NULL DEFAULT 'vi' AFTER slug;
ALTER TABLE articles ADD COLUMN translations JSON AFTER body;
