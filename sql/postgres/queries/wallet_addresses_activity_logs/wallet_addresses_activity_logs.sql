-- name: GetLogByWalletAddressID :many
SELECT * FROM wallet_addresses_activity_logs WHERE wallet_addresses_id = $1 ORDER BY created_at DESC LIMIT 100;

