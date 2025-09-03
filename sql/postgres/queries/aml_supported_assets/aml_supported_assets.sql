-- name: GetAllBySlug :many
SELECT sqlc.embed(c)
FROM aml_supported_assets ass
         INNER JOIN currencies c on c.id = ass.currency_id
WHERE service_slug = $1;

-- name: GetBySlugAndCurrencyID :one
SELECT sqlc.embed(c), sqlc.embed(ass)
FROM aml_supported_assets ass
         INNER JOIN currencies c on c.id = ass.currency_id
WHERE currency_id = $1
  AND service_slug = $2
LIMIT 1;