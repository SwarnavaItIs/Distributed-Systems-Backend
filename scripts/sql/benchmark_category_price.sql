DROP INDEX IF EXISTS idx_listings_category_price;

EXPLAIN ANALYZE
SELECT id, title, category_id, price_cents, status
FROM listings
WHERE category_id = 5
AND price_cents BETWEEN 1000000 AND 5000000
ORDER BY price_cents
LIMIT 20;

CREATE INDEX IF NOT EXISTS idx_listings_category_price
ON listings (category_id, price_cents);

EXPLAIN ANALYZE
SELECT id, title, category_id, price_cents, status
FROM listings
WHERE category_id = 5
AND price_cents BETWEEN 1000000 AND 5000000
ORDER BY price_cents
LIMIT 20;