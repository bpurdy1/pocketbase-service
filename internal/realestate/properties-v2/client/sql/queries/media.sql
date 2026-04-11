-- name: AddPhoto :one
INSERT INTO property_photos (
    property_id, listing_id, rental_id,
    source_url, local_path,
    caption, is_primary,
    width, height, size_bytes, mime_type, sort_order
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetPhoto :one
SELECT * FROM property_photos WHERE id = ? LIMIT 1;

-- name: ListPhotosForProperty :many
SELECT * FROM property_photos
WHERE property_id = ?
ORDER BY is_primary DESC, sort_order ASC, created_at ASC;

-- name: GetPrimaryPhotoForProperty :one
SELECT * FROM property_photos
WHERE property_id = ? AND is_primary = 1
LIMIT 1;

-- name: SetPrimaryPhoto :exec
UPDATE property_photos SET is_primary = 0 WHERE property_id = ?;

-- name: MarkPhotoAsPrimary :one
UPDATE property_photos SET is_primary = 1 WHERE id = ? RETURNING *;

-- name: DeletePhoto :exec
DELETE FROM property_photos WHERE id = ?;

-- name: DeletePhotosForProperty :exec
DELETE FROM property_photos WHERE property_id = ?;

-- name: UpsertListingSource :one
INSERT INTO listing_sources (
    property_id, source_name, source_url, source_type,
    last_seen_at, first_seen_at, is_active
) VALUES (?, ?, ?, ?, ?, ?, 1)
ON CONFLICT(property_id, source_url) DO UPDATE SET
    last_seen_at = excluded.last_seen_at,
    is_active    = 1
RETURNING *;

-- name: GetListingSources :many
SELECT * FROM listing_sources
WHERE property_id = ?
ORDER BY last_seen_at DESC;

-- name: DeactivateListingSource :exec
UPDATE listing_sources SET is_active = 0 WHERE id = ?;

-- name: ListRecentlySeenSources :many
SELECT ls.*, p.address, p.city, p.state
FROM listing_sources ls
JOIN properties p ON p.id = ls.property_id
WHERE ls.last_seen_at >= datetime('now', '-7 days')
  AND ls.is_active = 1
ORDER BY ls.last_seen_at DESC
LIMIT ? OFFSET ?;
