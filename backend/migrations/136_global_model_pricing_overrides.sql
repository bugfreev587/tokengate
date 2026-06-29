-- Global model pricing overrides used by the billing resolver.
-- These rows override provider fallback prices for an exact model name.

CREATE TABLE IF NOT EXISTS global_model_pricing_overrides (
    id                 BIGSERIAL      PRIMARY KEY,
    model              VARCHAR(255)   NOT NULL,
    provider           VARCHAR(50)    NOT NULL DEFAULT '',
    billing_mode       VARCHAR(20)    NOT NULL DEFAULT 'token',
    input_price        NUMERIC(20,12),
    output_price       NUMERIC(20,12),
    cache_write_price  NUMERIC(20,12),
    cache_read_price   NUMERIC(20,12),
    image_output_price NUMERIC(20,12),
    per_request_price  NUMERIC(20,12),
    created_at         TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_global_model_pricing_billing_mode
        CHECK (billing_mode IN ('token', 'per_request', 'image')),
    CONSTRAINT chk_global_model_pricing_non_negative
        CHECK (
            (input_price IS NULL OR input_price >= 0) AND
            (output_price IS NULL OR output_price >= 0) AND
            (cache_write_price IS NULL OR cache_write_price >= 0) AND
            (cache_read_price IS NULL OR cache_read_price >= 0) AND
            (image_output_price IS NULL OR image_output_price >= 0) AND
            (per_request_price IS NULL OR per_request_price >= 0)
        )
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_global_model_pricing_overrides_model_lower
    ON global_model_pricing_overrides (LOWER(model));

CREATE INDEX IF NOT EXISTS idx_global_model_pricing_overrides_provider
    ON global_model_pricing_overrides (provider);
