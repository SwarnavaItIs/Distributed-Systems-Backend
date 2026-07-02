INSERT INTO listings (
    seller_id,
    title,
    description,
    category_id,
    price_cents,
    status,
    created_at,
    updated_at
)
SELECT
    gen_random_uuid(),
    'Test Listing ' || gs,
    'Synthetic benchmark listing for DMB',
    ((gs % 20) + 1),
    ((1000 + (gs % 100000)) * 100),
    'ACTIVE',
    NOW() - (gs || ' minutes')::INTERVAL,
    NOW() - (gs || ' minutes')::INTERVAL
FROM generate_series(1, 100000) AS gs;