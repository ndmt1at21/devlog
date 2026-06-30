-- "Buy me a coffee" donation orders (Stripe Checkout + MoMo).
-- Amount is VND (zero-decimal). user_id/buyer fields are nullable (anonymous donations).

CREATE TABLE coffee_orders (
    id                CHAR(36) PRIMARY KEY,
    method            VARCHAR(20) NOT NULL,           -- "card" (Stripe) | "momo"
    amount            BIGINT NOT NULL,                -- VND, zero-decimal
    currency          VARCHAR(8) NOT NULL DEFAULT 'VND',
    status            VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending|completed|failed|cancelled
    buyer_name        VARCHAR(120),
    buyer_email       VARCHAR(255),
    user_id           VARCHAR(120),                   -- IAM sub when logged in
    stripe_session_id VARCHAR(255),
    momo_order_id     VARCHAR(120),
    momo_request_id   VARCHAR(120),
    created_at        DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    completed_at      DATETIME(6)
);

CREATE INDEX idx_coffee_status ON coffee_orders(status, created_at);
CREATE INDEX idx_coffee_user ON coffee_orders(user_id);
CREATE INDEX idx_coffee_stripe ON coffee_orders(stripe_session_id);
