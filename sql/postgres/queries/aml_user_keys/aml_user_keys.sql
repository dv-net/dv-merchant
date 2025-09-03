-- name: CreateUserKeys :exec
INSERT INTO aml_user_keys (key_id, user_id, value, created_at)
VALUES ($1, $2, $3, now())
ON CONFLICT (user_id, key_id) DO UPDATE set value      = $3,
                                            updated_at = now();

-- name: PrepareServiceDataByUserAndSlug :many
SElECT sqlc.embed(ams), ask.name, auk.value
FROM aml_user_keys auk
         INNER JOIN aml_service_keys ask ON auk.key_id = ask.id
         INNER JOIN aml_services ams ON ask.service_id = ams.id AND ams.slug = sqlc.arg(slug)
WHERE auk.user_id = sqlc.arg(user_id);


-- name: CreateOrUpdateUserKeys :one
INSERT INTO aml_user_keys (user_id, key_id, value, created_at)
VALUES (sqlc.arg(user_id), sqlc.arg(key_id), sqlc.arg(value), now())
ON CONFLICT (user_id, key_id) DO UPDATE
    SET value      = sqlc.arg(value),
        updated_at = now()
RETURNING *;

-- name: FetchAllBySlug :many
SELECT ask.id, ask.name, ask.description, auk.value
FROM aml_service_keys ask
         LEFT JOIN aml_user_keys auk ON ask.id = auk.key_id AND auk.user_id = $1
         INNER JOIN aml_services amls ON ask.service_id = amls.id
WHERE amls.slug = $2;

-- name: DeleteAllUserKeysBySlug :exec
DELETE
FROM aml_user_keys auk
    USING aml_service_keys ask
        JOIN aml_services s ON ask.service_id = s.id
WHERE auk.user_id = $1
  AND auk.key_id = ask.id
  AND s.slug = $2;

-- name: FetchServiceKeyIDBySlugAndName :one
SELECT ask.id
FROM aml_service_keys ask
         JOIN aml_services s ON ask.service_id = s.id
WHERE s.slug = $1 AND ask.name = $2;

-- name: DeleteAllUserKeysBySlugAndKeyID :exec
DELETE FROM aml_user_keys auk
    USING aml_service_keys ask
        JOIN aml_services s ON ask.service_id = s.id
WHERE auk.user_id = $1
  AND auk.key_id = ask.id
  AND s.slug = $2
  AND ask.id = $3;