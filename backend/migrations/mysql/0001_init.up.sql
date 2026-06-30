-- devlog blog initial schema (MySQL).
-- IDs are CHAR(36) UUIDv7 (app-side, time-ordered). List/object fields use JSON.
-- Timestamps are DATETIME(6); the app writes created_at/updated_at in UTC.

CREATE TABLE articles (
    id           CHAR(36) PRIMARY KEY,
    slug         VARCHAR(160) NOT NULL UNIQUE,
    ord          INT NOT NULL DEFAULT 0,
    featured     BOOLEAN NOT NULL DEFAULT FALSE,
    category     VARCHAR(80) NOT NULL,
    author       VARCHAR(120) NOT NULL,
    read_time    VARCHAR(40) NOT NULL,
    published_at DATETIME(6) NOT NULL,
    title        VARCHAR(300) NOT NULL,
    excerpt      TEXT NOT NULL,
    cover        VARCHAR(500),
    tags         JSON NOT NULL,
    series_slug  VARCHAR(80),
    series_part  INT,
    part_title   VARCHAR(200),
    body         JSON NOT NULL,
    created_at   DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at   DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6)
);

CREATE INDEX idx_articles_ord ON articles(ord);
CREATE INDEX idx_articles_category ON articles(category);
CREATE INDEX idx_articles_series ON articles(series_slug, series_part);

CREATE TABLE series (
    slug        VARCHAR(80) PRIMARY KEY,
    title       VARCHAR(200) NOT NULL,
    description TEXT NOT NULL,
    created_at  DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6)
);

CREATE TABLE comments (
    id           CHAR(36) PRIMARY KEY,
    article_slug VARCHAR(160) NOT NULL,
    name         VARCHAR(120) NOT NULL,
    body         TEXT NOT NULL,
    created_at   DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6)
);

CREATE INDEX idx_comments_article ON comments(article_slug, created_at);

CREATE TABLE subscriptions (
    id         CHAR(36) PRIMARY KEY,
    user_id    VARCHAR(120) NOT NULL UNIQUE,
    plan       VARCHAR(20) NOT NULL,
    status     VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6)
);
