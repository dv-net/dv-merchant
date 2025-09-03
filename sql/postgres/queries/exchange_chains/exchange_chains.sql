-- name: GetEnabledCurrencies :many
SELECT c.id, c.code, c.name, c.blockchain, ec.ticker, chain FROM exchange_chains ec
    LEFT JOIN currencies c on ec.currency_id = c.id
    WHERE slug = $1 AND c.status IS true;

-- name: GetCurrencyIDBySlugAndChain :one
SELECT ec.currency_id FROM exchange_chains ec
    WHERE slug = $1 AND chain = $2;

-- name: GetCurrencyIDByParams :one
SELECT ec.currency_id FROM exchange_chains ec
    WHERE slug = $1 AND chain = $2 AND ec.ticker = $3;

-- name: GetCurrencyIDByTicker :one
SELECT currency_id FROM exchange_chains WHERE ticker = $1;

-- name: GetTickerByCurrencyID :one
SELECT ticker FROM exchange_chains WHERE currency_id = $1 AND slug = $2;

-- name: GetAll :many
SELECT e.slug, e.currency_id, e.ticker, e.chain FROM exchange_chains e;