-- name: GetTransactionStatsByType :many
SELECT
    tr.type as transaction_type,
    tr.currency_id as currency_id,
    c.code as currency_code,
       COUNT(tr.id) as total_count,
       COALESCE(SUM(tr.amount), 0)::numeric as total_amount
FROM
    transactions tr
        JOIN currencies c ON tr.currency_id = c.id
WHERE
    tr.type IN ('transfer', 'deposit')
GROUP BY
    tr.type, tr.currency_id, c.code
ORDER BY
    tr.type;