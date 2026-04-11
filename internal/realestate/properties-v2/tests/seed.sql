-- Seed data for development / demo
-- Run after schema is applied.
-- Spatial shadow tables and R*Tree entries are populated automatically by triggers.

-- ─── Properties ─────────────────────────────────────────────────────────────

INSERT INTO properties (
    id, address, city, state, zip_code, county,
    lat, lng,
    property_name, property_type,
    bedrooms, bathrooms, sqft, building_sf, lot_sf,
    year_built, number_of_units, organization, notes
) VALUES
    ('p001', '1234 Sunset Blvd', 'Los Angeles', 'CA', '90028', 'Los Angeles',
     34.0983, -118.3267, 'Sunset Craftsman', 'sfr',
     3, 2.0, 1850, 1850, 6500, 1962, 1, 'Acme Realty', 'Corner lot, near Hollywood'),
    ('p002', '456 Market St', 'San Francisco', 'CA', '94105', 'San Francisco',
     37.7909, -122.3988, 'Market Loft', 'condo',
     2, 2.0, 1100, 1100, 0, 2005, 1, 'Bay Investments', 'High-rise condo, bay views'),
    ('p003', '789 Harbor Dr', 'San Diego', 'CA', '92101', 'San Diego',
     32.7141, -117.1731, 'Harbor Flats', 'mfr',
     0, 0.0, 8400, 8400, 12000, 1978, 6, 'Coast Holdings', '6-unit apartment building'),
    ('p004', '321 Oak Ave', 'Denver', 'CO', '80203', 'Denver',
     39.7285, -104.9817, 'Oak Victorian', 'sfr',
     4, 3.0, 2400, 2400, 7200, 1905, 1, 'Mile High RE', 'Historic Victorian, updated kitchen'),
    ('p005', '555 Lake Shore Dr', 'Chicago', 'IL', '60611', 'Cook',
     41.8975, -87.6220, 'Lakeshore Tower', 'condo',
     1, 1.0, 780, 780, 0, 1998, 1, 'Midwest Properties', 'Lake views from every room');

-- ─── Property Listings (active for-sale) ────────────────────────────────────

INSERT INTO property_listings (
    id, property_id, source_name, source_url, mls_id,
    list_price, price_per_sqft, status, days_on_market, expires_at
) VALUES
    ('l001', 'p001', 'Zillow', 'https://zillow.com/homes/p001', 'MLS-100001',
     849000, 458.92, 'active', 12, datetime('now', '+30 days')),
    ('l002', 'p002', 'Redfin', 'https://redfin.com/CA/SF/p002', 'MLS-100002',
     1195000, 1086.36, 'active', 5, datetime('now', '+30 days')),
    ('l003', 'p004', 'MLS', 'https://mls.com/listing/p004', 'MLS-100003',
     625000, 260.42, 'pending', 34, datetime('now', '+14 days'));

-- ─── Rental Listings (active rentals) ───────────────────────────────────────

INSERT INTO rental_listings (
    id, property_id, source_name, source_url, listing_ref,
    monthly_rent, security_deposit, rent_per_sqft,
    unit_number, bedrooms, bathrooms, sqft,
    available_date, lease_term,
    pets_allowed, furnished,
    status, days_on_market, expires_at
) VALUES
    ('r001', 'p003', 'Apartments.com', 'https://apartments.com/p003-unit1', 'APT-3001',
     2200, 2200, 1.05, '1', 1, 1.0, 1400, '2026-05-01', '12mo', 1, 0,
     'active', 8, datetime('now', '+30 days')),
    ('r002', 'p003', 'Craigslist', 'https://sd.craigslist.org/p003-unit3', 'CL-3003',
     2800, 2800, 1.33, '3', 2, 1.0, 1400, '2026-04-15', '12mo', 0, 0,
     'active', 3, datetime('now', '+30 days')),
    ('r003', 'p005', 'Zillow', 'https://zillow.com/rent/p005', 'ZR-5001',
     2400, 2400, 3.08, '', 1, 1.0, 780, '2026-05-01', 'month-to-month', 0, 1,
     'active', 15, datetime('now', '+30 days'));

-- ─── Property Photos ─────────────────────────────────────────────────────────

INSERT INTO property_photos (
    id, property_id, listing_id, rental_id,
    source_url, local_path, caption, is_primary,
    width, height, size_bytes, mime_type, sort_order
) VALUES
    ('ph001', 'p001', 'l001', NULL,
     'https://photos.zillow.com/p001-front.jpg', '', 'Front exterior', 1,
     1920, 1080, 524288, 'image/jpeg', 0),
    ('ph002', 'p001', 'l001', NULL,
     'https://photos.zillow.com/p001-kitchen.jpg', '', 'Updated kitchen', 0,
     1920, 1080, 487000, 'image/jpeg', 1),
    ('ph003', 'p002', 'l002', NULL,
     'https://photos.redfin.com/p002-living.jpg', '', 'Living room with bay views', 1,
     2048, 1536, 650000, 'image/jpeg', 0),
    ('ph004', 'p003', NULL, 'r001',
     'https://apartments.com/photos/p003-unit1.jpg', '', 'Unit 1 interior', 1,
     1280, 960, 312000, 'image/jpeg', 0);

-- ─── Listing Sources ─────────────────────────────────────────────────────────

INSERT INTO listing_sources (
    id, property_id, source_name, source_url, source_type, is_active
) VALUES
    ('s001', 'p001', 'Zillow',   'https://zillow.com/homes/p001', 'sale', 1),
    ('s002', 'p001', 'Redfin',   'https://redfin.com/CA/LA/p001', 'sale', 1),
    ('s003', 'p002', 'Redfin',   'https://redfin.com/CA/SF/p002', 'sale', 1),
    ('s004', 'p002', 'MLS',      'https://mls.com/listing/100002', 'sale', 1),
    ('s005', 'p003', 'Apartments.com', 'https://apartments.com/p003', 'rental', 1),
    ('s006', 'p005', 'Zillow',   'https://zillow.com/rent/p005',  'rental', 1);

-- ─── Historical Listings (archived) ──────────────────────────────────────────

INSERT INTO property_listings_history (
    id, listing_id, property_id,
    source_name, source_url, mls_id,
    list_price, price_per_sqft,
    status, days_on_market,
    listing_created_at, listing_expires_at,
    archive_reason
) VALUES
    ('lh001', 'l-old-001', 'p001', 'Zillow', 'https://zillow.com/homes/old-p001', 'MLS-099001',
     899000, 486.49, 'sold', 45,
     datetime('now', '-60 days'), datetime('now', '-30 days'),
     'sold'),
    ('lh002', 'l-old-002', 'p002', 'Redfin', 'https://redfin.com/CA/SF/old-p002', 'MLS-099002',
     1250000, 1136.36, 'off_market', 90,
     datetime('now', '-120 days'), datetime('now', '-90 days'),
     'expired');

-- ─── Historical Rentals (archived) ───────────────────────────────────────────

INSERT INTO rental_listings_history (
    id, rental_id, property_id,
    source_name, source_url, listing_ref,
    unit_number, monthly_rent, security_deposit,
    bedrooms, bathrooms, sqft,
    available_date, lease_term, status, days_on_market,
    listing_created_at, listing_expires_at,
    archive_reason
) VALUES
    ('rh001', 'r-old-001', 'p003', 'Craigslist', 'https://sd.craigslist.org/old-3003', 'CL-OLD-1',
     '2', 2600, 2600, 2, 1.0, 1400, '2025-09-01', '12mo', 'rented', 7,
     datetime('now', '-120 days'), datetime('now', '-90 days'),
     'rented');
