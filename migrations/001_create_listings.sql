CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS listings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    seller_id UUID NOT NULL,

    title VARCHAR(120) NOT NULL,
    description TEXT,

    category_id BIGINT NOT NULL,

    price_cents BIGINT NOT NULL CHECK (price_cents > 0),

    status VARCHAR(30) NOT NULL DEFAULT 'ACTIVE',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_listings_category_price
ON listings (category_id, price_cents);

CREATE INDEX IF NOT EXISTS idx_listings_seller_id
ON listings (seller_id);

CREATE INDEX IF NOT EXISTS idx_listings_created_at
ON listings (created_at);