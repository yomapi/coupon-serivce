echo "✅ Step 0: Postgres 테이블 생성"

docker exec -i local-postgres psql -U postgres -d coupon_service <<EOF
CREATE TABLE IF NOT EXISTS campaigns (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    coupon_count INT NOT NULL,
    start_at TIMESTAMP NOT NULL,
    end_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS coupons (
    id SERIAL PRIMARY KEY,
    code TEXT NOT NULL,
    user_id INT NOT NULL,
    campaign_id INT NOT NULL,
    issued_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (user_id, campaign_id)
);
EOF