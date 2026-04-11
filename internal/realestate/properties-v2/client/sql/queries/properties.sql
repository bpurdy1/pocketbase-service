-- name: CreateProperty :one
INSERT INTO properties (
    address, city, state, zip_code, county,
    lat, lng,
    property_name, property_type,
    bedrooms, bathrooms, sqft, building_sf, lot_sf,
    year_built, number_of_units,
    organization, notes
) VALUES (
    ?, ?, ?, ?, ?,
    ?, ?,
    ?, ?,
    ?, ?, ?, ?, ?,
    ?, ?,
    ?, ?
)
RETURNING *;

-- name: GetProperty :one
SELECT * FROM properties WHERE id = ? LIMIT 1;

-- name: GetPropertyByAddress :one
SELECT * FROM properties WHERE address = ? LIMIT 1;

-- name: ListProperties :many
SELECT * FROM properties
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListPropertiesByOrg :many
SELECT * FROM properties
WHERE organization = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListPropertiesByType :many
SELECT * FROM properties
WHERE property_type = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListPropertiesByCityState :many
SELECT * FROM properties
WHERE city = ? AND state = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: UpdateProperty :one
UPDATE properties SET
    address        = ?,
    city           = ?,
    state          = ?,
    zip_code       = ?,
    county         = ?,
    lat            = ?,
    lng            = ?,
    property_name  = ?,
    property_type  = ?,
    bedrooms       = ?,
    bathrooms      = ?,
    sqft           = ?,
    building_sf    = ?,
    lot_sf         = ?,
    year_built     = ?,
    number_of_units = ?,
    organization   = ?,
    notes          = ?
WHERE id = ?
RETURNING *;

-- name: UpdatePropertyCoords :one
UPDATE properties SET lat = ?, lng = ? WHERE id = ? RETURNING *;

-- name: DeleteProperty :exec
DELETE FROM properties WHERE id = ?;

-- name: CountProperties :one
SELECT COUNT(*) FROM properties;

-- name: GetPropertyWithListingAndRental :one
SELECT
    p.*,
    pl.id           AS listing_id,
    pl.list_price,
    pl.status       AS listing_status,
    pl.source_url   AS listing_url,
    pl.expires_at   AS listing_expires_at,
    rl.id           AS rental_id,
    rl.monthly_rent,
    rl.status       AS rental_status,
    rl.source_url   AS rental_url,
    rl.expires_at   AS rental_expires_at
FROM properties p
LEFT JOIN property_listings pl ON pl.property_id = p.id
LEFT JOIN rental_listings   rl ON rl.property_id = p.id
WHERE p.id = ?
LIMIT 1;
