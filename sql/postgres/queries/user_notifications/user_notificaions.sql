-- name: GetUserNotificationChannels :one
SELECT CASE WHEN n.category IS NULL THEN false ELSE coalesce(un.tg_enabled, false) END::bool AS tg_enabled,
       CASE WHEN n.category IS NULL THEN true ELSE coalesce(un.email_enabled, false) END::bool AS email_enabled
FROM notifications n
         LEFT JOIN user_notifications un on un.notification_id = n.id AND un.user_id = $1
WHERE n.type = $2
LIMIT 1;

-- name: GetUserListWithCategory :many
SELECT n.id,
       COALESCE(n.category, '')::varchar as category,
       n.type,
       coalesce(un.tg_enabled, false),
       coalesce(un.email_enabled, CASE when n.category = 'system' THEN true ELSE false END)
FROM notifications n
         LEFT JOIN user_notifications un ON un.notification_id = n.id AND un.user_id = $1 AND n.category IS NOT NULL
WHERE n.category IS NOT NULL;

-- name: GetByUserAndID :one
SELECT sqlc.embed(un), n.category
from user_notifications un
         INNER JOIN notifications n ON un.notification_id = n.id
WHERE un.id = $1
  and un.user_id = $2
LIMIT 1;

-- name: CreateOrUpdate :batchexec
INSERT INTO user_notifications (user_id, notification_id, email_enabled, tg_enabled, created_at)
VALUES ($1, $2, $3, $4, now())
ON CONFLICT (user_id, notification_id) DO UPDATE set tg_enabled    = $4,
                                                     email_enabled = $3,
                                                     updated_at    = now();