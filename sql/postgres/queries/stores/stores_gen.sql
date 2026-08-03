-- name: Create :one
INSERT INTO stores (user_id, name, site, currency_id, rate_source, return_url, success_url, status, created_at, public_payment_form_enabled, description, verification_comment)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now(), $9, $10, $11)
	RETURNING *;

-- name: GetByID :one
SELECT * FROM stores WHERE id=$1 LIMIT 1;

-- name: Update :one
UPDATE stores
	SET name=$1, site=$2, currency_id=$3, rate_source=$4, return_url=$5, success_url=$6, rate_scale=$7, status=$8, minimal_payment=$9, updated_at=now(), public_payment_form_enabled=$10, description=$11
WHERE id=$12
	RETURNING *;
