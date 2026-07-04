-- Add the article author's stable user id so edit authorization keys off
-- identity rather than the mutable, non-unique display name. Nullable on
-- purpose: seed/demo articles have no owning account and stay read-only
-- (a NULL author_id can never match a signed-in user's id).

ALTER TABLE articles ADD COLUMN author_id CHAR(36) NULL AFTER author;
