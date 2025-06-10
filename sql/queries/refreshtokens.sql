-- name: CreateRefreshToken :one
INSERT INTO refreshtokens (token, created_at, updated_at, user_id, expires_at)
VALUES (
    $1, NOW(), NOW(), $2, $3
)
RETURNING *;

-- name: GetRefreshTokenByToken :one
SELECT * FROM refreshtokens WHERE token = $1;

-- name: SetRevokeRefreshToken :exec
UPDATE refreshtokens
SET updated_at = NOW(),
revoked_at = NOW()
WHERE token = $1;
