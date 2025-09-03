-- name: GetOneByHashAndBlockchain :one
SELECT * FROM unconfirmed_transactions WHERE tx_hash=$1 and blockchain=$2 LIMIT 1;

-- name: CollapseAllByConfirmed :exec
DELETE FROM unconfirmed_transactions ut
       USING transactions t
       WHERE t.tx_hash=ut.tx_hash AND t.currency_id=ut.currency_id AND t.bc_uniq_key=ut.bc_uniq_key;