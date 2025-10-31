-- name: GetByInvoiceID :many
SELECT *
FROM invoice_addresses
         LEFT JOIN wallet_addresses ON invoice_addresses.wallet_address_id = wallet_addresses.id
WHERE invoice_id = $1;

-- name: GetByIDWithAddress :one
SELECT *
FROM invoice_addresses
         LEFT JOIN wallet_addresses ON invoice_addresses.wallet_address_id = wallet_addresses.id
WHERE invoice_addresses.id = $1 LIMIT 1;

-- name: CreateWithWalletAddress :one
WITH inserted AS (
    INSERT INTO invoice_addresses (invoice_id, wallet_address_id, rate_at_creation, created_at)
        VALUES ($1, $2, $3, now())
        RETURNING *
)
SELECT
    inserted.*,
    wallet_addresses.*
FROM inserted
         LEFT JOIN wallet_addresses ON inserted.wallet_address_id = wallet_addresses.id;

-- name: GetInvoiceAddressByInvoiceAndCurrency :one
SELECT *
FROM invoice_addresses
         LEFT JOIN wallet_addresses ON invoice_addresses.wallet_address_id = wallet_addresses.id
WHERE invoice_id = $1 AND wallet_addresses.currency_id = $2 LIMIT 1;

-- name: GetByWalletAddressID :one
SELECT *
FROM invoice_addresses
         LEFT JOIN invoices ON invoice_addresses.invoice_id = invoices.id
WHERE invoice_addresses.wallet_address_id = $1 ORDER BY invoices.created_at DESC LIMIT 1;
