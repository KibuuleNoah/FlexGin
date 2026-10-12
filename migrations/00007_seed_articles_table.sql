-- +goose Up
-- +goose StatementBegin
INSERT INTO articles (title, slug, body, author_id, published)
SELECT
	seed.title,
	seed.slug,
	seed.body,
	u.id,
	true
FROM (
	VALUES
		('Getting Started with Escrow Payments', 'getting-started-with-escrow-payments', 'A walkthrough of how escrow transactions protect both buyers and sellers in B2B deals, from initiation to release.', 0),
		('Why API Gateways Matter for B2B Platforms', 'why-api-gateways-matter-for-b2b-platforms', 'API gateways centralize auth, rate limiting, and routing across microservices, reducing duplicated logic at every edge.', 1),
		('Designing Idempotent Payment Endpoints', 'designing-idempotent-payment-endpoints', 'Idempotency keys prevent duplicate charges when clients retry failed requests, a must for any payment-critical system.', 2),
		('A Guide to Postgres Indexing Strategies', 'a-guide-to-postgres-indexing-strategies', 'Covers B-tree vs GIN indexes, composite index ordering, and when an index actually helps versus adds write overhead.', 0),
		('Structuring Go Services for Testability', 'structuring-go-services-for-testability', 'Separating service, controller, and repository layers makes unit testing business logic possible without a live database.', 1)
) AS seed(title, slug, body, row_offset)
JOIN (
	SELECT id, ROW_NUMBER() OVER (ORDER BY created_at) - 1 AS rn
	FROM users
) u ON u.rn = seed.row_offset % (SELECT COUNT(*) FROM users);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM articles
WHERE slug IN (
	'getting-started-with-escrow-payments',
	'why-api-gateways-matter-for-b2b-platforms',
	'designing-idempotent-payment-endpoints',
	'a-guide-to-postgres-indexing-strategies',
	'structuring-go-services-for-testability'
);
-- +goose StatementEnd
