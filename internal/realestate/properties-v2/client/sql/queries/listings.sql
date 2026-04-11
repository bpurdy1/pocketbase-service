-- name: CreateListing :one
INSERT INTO property_listings (
    property_id, source_name, source_url, mls_id,
    list_price, price_per_sqft,
    status, days_on_market, expires_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetListing :one
SELECT * FROM property_listings WHERE id = ? LIMIT 1;

-- name: GetActiveListingByProperty :one
SELECT * FROM property_listings
WHERE property_id = ?
LIMIT 1;

-- name: ListActiveListings :many
SELECT pl.*, p.address, p.city, p.state, p.zip_code, p.lat, p.lng,
       p.bedrooms, p.bathrooms, p.sqft, p.property_type
FROM property_listings pl
JOIN properties p ON p.id = pl.property_id
WHERE pl.status = 'active'
  AND pl.expires_at > datetime('now')
ORDER BY pl.created_at DESC
LIMIT ? OFFSET ?;

-- name: ListExpiredListings :many
SELECT * FROM property_listings
WHERE expires_at <= datetime('now')
ORDER BY expires_at ASC;

-- name: UpdateListingTTL :one
UPDATE property_listings
SET expires_at = ?
WHERE id = ?
RETURNING *;

-- name: UpdateListingStatus :one
UPDATE property_listings SET status = ? WHERE id = ? RETURNING *;

-- name: UpdateListing :one
UPDATE property_listings SET
    source_name     = ?,
    source_url      = ?,
    mls_id          = ?,
    list_price      = ?,
    price_per_sqft  = ?,
    status          = ?,
    days_on_market  = ?,
    expires_at      = ?
WHERE id = ?
RETURNING *;

-- name: DeleteListing :exec
DELETE FROM property_listings WHERE id = ?;

-- name: ArchiveListing :one
INSERT INTO property_listings_history (
    listing_id, property_id,
    source_name, source_url, mls_id,
    list_price, price_per_sqft,
    status, days_on_market,
    listing_created_at, listing_expires_at,
    archive_reason
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetListingHistory :many
SELECT * FROM property_listings_history
WHERE property_id = ?
ORDER BY archived_at DESC
LIMIT ? OFFSET ?;

-- name: GetListingHistoryByListingID :many
SELECT * FROM property_listings_history
WHERE listing_id = ?
ORDER BY archived_at DESC;

-- name: ListAllListingHistory :many
SELECT plh.*, p.address, p.city, p.state
FROM property_listings_history plh
JOIN properties p ON p.id = plh.property_id
ORDER BY plh.archived_at DESC
LIMIT ? OFFSET ?;
