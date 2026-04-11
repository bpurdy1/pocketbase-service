-- name: GetPropertyHistory :many
SELECT * FROM properties_history
WHERE property_id = ?
ORDER BY changed_at DESC
LIMIT ? OFFSET ?;

-- name: GetPropertyHistoryByChangeType :many
SELECT * FROM properties_history
WHERE property_id = ? AND change_type = ?
ORDER BY changed_at DESC;

-- name: ListAllPropertyHistory :many
SELECT * FROM properties_history
ORDER BY changed_at DESC
LIMIT ? OFFSET ?;
