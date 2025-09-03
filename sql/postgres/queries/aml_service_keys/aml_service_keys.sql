-- name: GetAllByServiceSlug :many
SELECT * FROM aml_service_keys ask INNER JOIN aml_services amls ON ask.service_id = amls.id AND amls.slug = sqlc.arg(slug);