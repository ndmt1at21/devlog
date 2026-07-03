-- Per-user article reactions: likes and bookmarks. One row per (article, user,
-- kind); the composite PK makes toggling idempotent. user_id matches
-- subscriptions.user_id (IAM subject), article_slug matches articles.slug.

CREATE TABLE reactions (
    article_slug VARCHAR(160) NOT NULL,
    user_id      VARCHAR(120) NOT NULL,
    kind         VARCHAR(20) NOT NULL,
    created_at   DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (article_slug, user_id, kind)
);

-- "My bookmarks" listing: newest saved first.
CREATE INDEX idx_reactions_user ON reactions(user_id, kind, created_at);
