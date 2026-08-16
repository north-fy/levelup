CREATE TABLE shop_items (
    id          BIGSERIAL PRIMARY KEY,
    seller_id   BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    title       VARCHAR(255) NOT NULL,
    description VARCHAR(1024) NOT NULL DEFAULT '',
    price_gold  INTEGER NOT NULL,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_shop_items_seller_id ON shop_items (seller_id);
CREATE INDEX idx_shop_items_active ON shop_items (is_active);

CREATE TABLE purchases (
    id         BIGSERIAL PRIMARY KEY,
    item_id    BIGINT NOT NULL REFERENCES shop_items (id) ON DELETE CASCADE,
    buyer_id   BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    seller_id  BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    price      INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_purchases_buyer_id ON purchases (buyer_id);
CREATE INDEX idx_purchases_seller_id ON purchases (seller_id);