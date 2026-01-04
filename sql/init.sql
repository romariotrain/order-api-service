CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE orders (
                        id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
                        status text NOT NULL,
                        version bigint NOT NULL DEFAULT 0,

                        fail_reason_code text NULL,
                        fail_reason_detail text NULL,

                        created_at timestamptz NOT NULL DEFAULT now(),
                        updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_orders_status ON orders(status);