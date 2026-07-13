-- Local accounts + refresh tokens for the embedded auth provider. Schema
-- mirrors the IAM service's users/refresh_tokens tables, single-tenant.

CREATE TABLE users (
    id             CHAR(36) PRIMARY KEY,
    email          VARCHAR(254) NOT NULL UNIQUE,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    name           VARCHAR(255) NOT NULL DEFAULT '',
    password_hash  VARCHAR(255),
    role           VARCHAR(32) NOT NULL DEFAULT 'reader',
    google_sub     VARCHAR(255),
    created_at     DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at     DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    UNIQUE KEY uq_users_google_sub (google_sub)
);

CREATE TABLE refresh_tokens (
    id         CHAR(36) PRIMARY KEY,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    user_id    CHAR(36) NOT NULL,
    expires_at DATETIME(6) NOT NULL,
    revoked_at DATETIME(6),
    CONSTRAINT fk_refresh_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_refresh_user ON refresh_tokens(user_id);
