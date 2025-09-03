-- name: GetCurrenciesHasBalance :many
SELECT *
FROM currencies
where status = true
  and has_balance = true;

-- name: GetCurrencyWithBalanceByBlockchain :one
SELECT *
FROM currencies
WHERE blockchain = sqlc.arg(blockchain)
  AND has_balance = true
LIMIT 1;

-- name: GetCurrenciesByBlockchain :many
SELECT *
FROM currencies
WHERE blockchain = sqlc.arg(blockchain);

-- name: GetCurrenciesEnabled :many
SELECT *
FROM currencies
WHERE status = true
ORDER BY blockchain, code;

-- name: GetEnabledCurrencyById :one
SELECT *
FROM currencies
where id = $1 and status = true
limit 1;

-- name: GetCurrencyByBlockchainAndContract :one
SELECT *
FROM currencies
WHERE blockchain = sqlc.arg(blockchain)
  AND (contract_address = $1 OR (contract_address IS NULL AND code = UPPER($1)))
  AND status = true
LIMIT 1;

-- name: GetEnabledCurrencyByCode :one
SELECT *
FROM currencies
WHERE code = $1
  AND status = true
  AND blockchain = $2
LIMIT 1;