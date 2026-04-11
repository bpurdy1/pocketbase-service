-- name: CreateRental :one
INSERT INTO rental_listings (
    property_id, source_name, source_url, listing_ref,
    monthly_rent, security_deposit, rent_per_sqft,
    unit_number, bedrooms, bathrooms, sqft,
    available_date, lease_term,
    pets_allowed, furnished,
    status, days_on_market, expires_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetRental :one
SELECT * FROM rental_listings WHERE id = ? LIMIT 1;

-- name: GetActiveRentalByPropertyUnit :one
SELECT * FROM rental_listings
WHERE property_id = ? AND unit_number = ?
LIMIT 1;

-- name: ListActiveRentals :many
SELECT rl.*, p.address, p.city, p.state, p.zip_code, p.lat, p.lng,
       p.property_type, p.year_built
FROM rental_listings rl
JOIN properties p ON p.id = rl.property_id
WHERE rl.status = 'active'
  AND rl.expires_at > datetime('now')
ORDER BY rl.created_at DESC
LIMIT ? OFFSET ?;

-- name: ListActiveRentalsByPriceRange :many
SELECT rl.*, p.address, p.city, p.state, p.zip_code, p.lat, p.lng
FROM rental_listings rl
JOIN properties p ON p.id = rl.property_id
WHERE rl.status = 'active'
  AND rl.expires_at > datetime('now')
  AND rl.monthly_rent >= ?
  AND rl.monthly_rent <= ?
ORDER BY rl.monthly_rent ASC
LIMIT ? OFFSET ?;

-- name: ListActiveRentalsByBedrooms :many
SELECT rl.*, p.address, p.city, p.state, p.zip_code, p.lat, p.lng
FROM rental_listings rl
JOIN properties p ON p.id = rl.property_id
WHERE rl.status = 'active'
  AND rl.expires_at > datetime('now')
  AND rl.bedrooms >= ?
ORDER BY rl.monthly_rent ASC
LIMIT ? OFFSET ?;

-- name: ListExpiredRentals :many
SELECT * FROM rental_listings
WHERE expires_at <= datetime('now')
ORDER BY expires_at ASC;

-- name: UpdateRentalTTL :one
UPDATE rental_listings
SET expires_at = ?
WHERE id = ?
RETURNING *;

-- name: UpdateRentalStatus :one
UPDATE rental_listings SET status = ? WHERE id = ? RETURNING *;

-- name: UpdateRental :one
UPDATE rental_listings SET
    source_name      = ?,
    source_url       = ?,
    listing_ref      = ?,
    monthly_rent     = ?,
    security_deposit = ?,
    rent_per_sqft    = ?,
    unit_number      = ?,
    bedrooms         = ?,
    bathrooms        = ?,
    sqft             = ?,
    available_date   = ?,
    lease_term       = ?,
    pets_allowed     = ?,
    furnished        = ?,
    status           = ?,
    days_on_market   = ?,
    expires_at       = ?
WHERE id = ?
RETURNING *;

-- name: DeleteRental :exec
DELETE FROM rental_listings WHERE id = ?;

-- name: ArchiveRental :one
INSERT INTO rental_listings_history (
    rental_id, property_id,
    source_name, source_url, listing_ref,
    unit_number, monthly_rent, security_deposit,
    bedrooms, bathrooms, sqft,
    available_date, lease_term,
    status, days_on_market,
    listing_created_at, listing_expires_at,
    archive_reason
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetRentalHistory :many
SELECT * FROM rental_listings_history
WHERE property_id = ?
ORDER BY archived_at DESC
LIMIT ? OFFSET ?;

-- name: GetRentalHistoryByRentalID :many
SELECT * FROM rental_listings_history
WHERE rental_id = ?
ORDER BY archived_at DESC;
