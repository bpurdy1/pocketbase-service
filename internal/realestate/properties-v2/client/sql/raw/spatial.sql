-- Spatial / R*Tree queries
-- All bounding box params: minLat, maxLat, minLng, maxLng

-- name: FindPropertiesInBBox :many
-- Find all properties whose coordinates fall within a lat/lng bounding box.
SELECT p.*
FROM property_rtree r
JOIN property_spatial_map m ON m.rid = r.id
JOIN properties p ON p.id = m.property_id
WHERE r.min_lat >= ?   -- minLat
  AND r.max_lat <= ?   -- maxLat
  AND r.min_lng >= ?   -- minLng
  AND r.max_lng <= ?   -- maxLng
ORDER BY p.created_at DESC;

-- name: FindActiveListingsInBBox :many
-- Find active for-sale listings within a bounding box.
SELECT p.*, pl.list_price, pl.status, pl.source_url, pl.expires_at, pl.mls_id
FROM listing_rtree r
JOIN listing_spatial_map m ON m.rid = r.id
JOIN property_listings pl ON pl.id = m.listing_id
JOIN properties p ON p.id = pl.property_id
WHERE r.min_lat >= ?   -- minLat
  AND r.max_lat <= ?   -- maxLat
  AND r.min_lng >= ?   -- minLng
  AND r.max_lng <= ?   -- maxLng
  AND pl.status = 'active'
  AND pl.expires_at > datetime('now')
ORDER BY pl.list_price ASC;

-- name: FindActiveListingsInBBoxByPrice :many
-- Active sale listings in a bounding box filtered by price range.
SELECT p.*, pl.list_price, pl.status, pl.source_url, pl.expires_at, pl.mls_id
FROM listing_rtree r
JOIN listing_spatial_map m ON m.rid = r.id
JOIN property_listings pl ON pl.id = m.listing_id
JOIN properties p ON p.id = pl.property_id
WHERE r.min_lat >= ?   -- minLat
  AND r.max_lat <= ?   -- maxLat
  AND r.min_lng >= ?   -- minLng
  AND r.max_lng <= ?   -- maxLng
  AND pl.status = 'active'
  AND pl.expires_at > datetime('now')
  AND pl.list_price BETWEEN ? AND ?   -- minPrice, maxPrice
ORDER BY pl.list_price ASC;

-- name: FindActiveRentalsInBBox :many
-- Find active rental listings within a bounding box.
SELECT p.*, rl.monthly_rent, rl.bedrooms, rl.bathrooms, rl.sqft,
       rl.unit_number, rl.available_date, rl.source_url, rl.expires_at
FROM rental_rtree r
JOIN rental_spatial_map m ON m.rid = r.id
JOIN rental_listings rl ON rl.id = m.rental_id
JOIN properties p ON p.id = rl.property_id
WHERE r.min_lat >= ?   -- minLat
  AND r.max_lat <= ?   -- maxLat
  AND r.min_lng >= ?   -- minLng
  AND r.max_lng <= ?   -- maxLng
  AND rl.status = 'active'
  AND rl.expires_at > datetime('now')
ORDER BY rl.monthly_rent ASC;

-- name: FindActiveRentalsInBBoxByRent :many
-- Active rentals in bounding box filtered by rent range and bedroom count.
SELECT p.*, rl.monthly_rent, rl.bedrooms, rl.bathrooms, rl.sqft,
       rl.unit_number, rl.available_date, rl.source_url, rl.expires_at
FROM rental_rtree r
JOIN rental_spatial_map m ON m.rid = r.id
JOIN rental_listings rl ON rl.id = m.rental_id
JOIN properties p ON p.id = rl.property_id
WHERE r.min_lat >= ?    -- minLat
  AND r.max_lat <= ?    -- maxLat
  AND r.min_lng >= ?    -- minLng
  AND r.max_lng <= ?    -- maxLng
  AND rl.status = 'active'
  AND rl.expires_at > datetime('now')
  AND rl.monthly_rent BETWEEN ? AND ?   -- minRent, maxRent
  AND rl.bedrooms >= ?                  -- minBedrooms
ORDER BY rl.monthly_rent ASC;

-- name: FindPropertiesNearPoint :many
-- Approximate radius search: compute a bounding box from a center point
-- and radius in degrees (~0.009 deg per km). For precise haversine distance,
-- post-filter in Go after retrieving candidates.
SELECT p.*,
       ((p.lat - ?) * (p.lat - ?) + (p.lng - ?) * (p.lng - ?)) AS dist_sq
FROM property_rtree r
JOIN property_spatial_map m ON m.rid = r.id
JOIN properties p ON p.id = m.property_id
WHERE r.min_lat >= ? - ?   -- centerLat - radiusDeg
  AND r.max_lat <= ? + ?   -- centerLat + radiusDeg
  AND r.min_lng >= ? - ?   -- centerLng - radiusDeg
  AND r.max_lng <= ? + ?   -- centerLng + radiusDeg
ORDER BY dist_sq ASC
LIMIT ?;
