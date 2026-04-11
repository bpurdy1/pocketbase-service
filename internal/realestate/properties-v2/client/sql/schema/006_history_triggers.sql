-- Automatic history tracking via SQLite triggers.
--
-- Properties:  every UPDATE and DELETE saves the OLD row to properties_history.
-- Listings:    every UPDATE saves old row; DELETE auto-archives if not already done.
-- Rentals:     same pattern as listings.

-- ── properties_history ────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS properties_history (
    id             TEXT PRIMARY KEY NOT NULL DEFAULT ('ph'||lower(hex(randomblob(7)))),
    -- snapshot of the property row BEFORE the change
    property_id    TEXT NOT NULL,
    address        TEXT NOT NULL DEFAULT '',
    city           TEXT NOT NULL DEFAULT '',
    state          TEXT NOT NULL DEFAULT '',
    zip_code       TEXT NOT NULL DEFAULT '',
    county         TEXT NOT NULL DEFAULT '',
    lat            REAL NOT NULL DEFAULT 0,
    lng            REAL NOT NULL DEFAULT 0,
    property_name  TEXT NOT NULL DEFAULT '',
    property_type  TEXT NOT NULL DEFAULT '',
    bedrooms       REAL NOT NULL DEFAULT 0,
    bathrooms      REAL NOT NULL DEFAULT 0,
    sqft           REAL NOT NULL DEFAULT 0,
    building_sf    REAL NOT NULL DEFAULT 0,
    lot_sf         REAL NOT NULL DEFAULT 0,
    year_built     INTEGER NOT NULL DEFAULT 0,
    number_of_units INTEGER NOT NULL DEFAULT 0,
    organization   TEXT NOT NULL DEFAULT '',
    notes          TEXT NOT NULL DEFAULT '',
    row_created_at TEXT NOT NULL DEFAULT '',
    row_updated_at TEXT NOT NULL DEFAULT '',
    -- audit fields
    change_type    TEXT NOT NULL DEFAULT '',  -- 'updated' | 'deleted'
    changed_at     TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_properties_history_property
    ON properties_history(property_id);

CREATE INDEX IF NOT EXISTS idx_properties_history_changed
    ON properties_history(changed_at);

-- Save old row on UPDATE of meaningful fields only.
-- Excludes updated_at so the auto-timestamp trigger doesn't cause a second snapshot.
CREATE TRIGGER IF NOT EXISTS trg_properties_history_update
AFTER UPDATE OF address, city, state, zip_code, county,
               lat, lng, property_name, property_type,
               bedrooms, bathrooms, sqft, building_sf, lot_sf,
               year_built, number_of_units, organization, notes
ON properties
FOR EACH ROW
BEGIN
    INSERT INTO properties_history (
        property_id, address, city, state, zip_code, county,
        lat, lng, property_name, property_type,
        bedrooms, bathrooms, sqft, building_sf, lot_sf,
        year_built, number_of_units, organization, notes,
        row_created_at, row_updated_at, change_type
    ) VALUES (
        OLD.id, OLD.address, OLD.city, OLD.state, OLD.zip_code, OLD.county,
        OLD.lat, OLD.lng, OLD.property_name, OLD.property_type,
        OLD.bedrooms, OLD.bathrooms, OLD.sqft, OLD.building_sf, OLD.lot_sf,
        OLD.year_built, OLD.number_of_units, OLD.organization, OLD.notes,
        OLD.created_at, OLD.updated_at, 'updated'
    );
END;

-- Save old row on DELETE
CREATE TRIGGER IF NOT EXISTS trg_properties_history_delete
AFTER DELETE ON properties
FOR EACH ROW
BEGIN
    INSERT INTO properties_history (
        property_id, address, city, state, zip_code, county,
        lat, lng, property_name, property_type,
        bedrooms, bathrooms, sqft, building_sf, lot_sf,
        year_built, number_of_units, organization, notes,
        row_created_at, row_updated_at, change_type
    ) VALUES (
        OLD.id, OLD.address, OLD.city, OLD.state, OLD.zip_code, OLD.county,
        OLD.lat, OLD.lng, OLD.property_name, OLD.property_type,
        OLD.bedrooms, OLD.bathrooms, OLD.sqft, OLD.building_sf, OLD.lot_sf,
        OLD.year_built, OLD.number_of_units, OLD.organization, OLD.notes,
        OLD.created_at, OLD.updated_at, 'deleted'
    );
END;


-- ── property_listings auto-archive ───────────────────────────────────────────

-- Save old listing row on UPDATE of meaningful fields.
-- Excludes updated_at to avoid double-snapshot from the auto-timestamp trigger.
CREATE TRIGGER IF NOT EXISTS trg_listing_history_update
AFTER UPDATE OF source_name, source_url, mls_id,
               list_price, price_per_sqft,
               status, days_on_market, expires_at
ON property_listings
FOR EACH ROW
BEGIN
    INSERT INTO property_listings_history (
        listing_id, property_id,
        source_name, source_url, mls_id,
        list_price, price_per_sqft,
        status, days_on_market,
        listing_created_at, listing_expires_at,
        archive_reason
    ) VALUES (
        OLD.id, OLD.property_id,
        OLD.source_name, OLD.source_url, OLD.mls_id,
        OLD.list_price, OLD.price_per_sqft,
        OLD.status, OLD.days_on_market,
        OLD.created_at, OLD.expires_at,
        'updated'
    );
END;

-- Auto-archive on DELETE (catches cases where app skips manual archive)
CREATE TRIGGER IF NOT EXISTS trg_listing_history_delete
BEFORE DELETE ON property_listings
FOR EACH ROW
BEGIN
    INSERT OR IGNORE INTO property_listings_history (
        listing_id, property_id,
        source_name, source_url, mls_id,
        list_price, price_per_sqft,
        status, days_on_market,
        listing_created_at, listing_expires_at,
        archive_reason
    ) VALUES (
        OLD.id, OLD.property_id,
        OLD.source_name, OLD.source_url, OLD.mls_id,
        OLD.list_price, OLD.price_per_sqft,
        OLD.status, OLD.days_on_market,
        OLD.created_at, OLD.expires_at,
        'deleted'
    );
END;


-- ── rental_listings auto-archive ─────────────────────────────────────────────

-- Save old rental row on UPDATE of meaningful fields.
-- Excludes updated_at to avoid double-snapshot from the auto-timestamp trigger.
CREATE TRIGGER IF NOT EXISTS trg_rental_history_update
AFTER UPDATE OF source_name, source_url, listing_ref,
               monthly_rent, security_deposit, rent_per_sqft,
               unit_number, bedrooms, bathrooms, sqft,
               available_date, lease_term, pets_allowed, furnished,
               status, days_on_market, expires_at
ON rental_listings
FOR EACH ROW
BEGIN
    INSERT INTO rental_listings_history (
        rental_id, property_id,
        source_name, source_url, listing_ref,
        unit_number, monthly_rent, security_deposit,
        bedrooms, bathrooms, sqft,
        available_date, lease_term,
        status, days_on_market,
        listing_created_at, listing_expires_at,
        archive_reason
    ) VALUES (
        OLD.id, OLD.property_id,
        OLD.source_name, OLD.source_url, OLD.listing_ref,
        OLD.unit_number, OLD.monthly_rent, OLD.security_deposit,
        OLD.bedrooms, OLD.bathrooms, OLD.sqft,
        OLD.available_date, OLD.lease_term,
        OLD.status, OLD.days_on_market,
        OLD.created_at, OLD.expires_at,
        'updated'
    );
END;

-- Auto-archive on DELETE
CREATE TRIGGER IF NOT EXISTS trg_rental_history_delete
BEFORE DELETE ON rental_listings
FOR EACH ROW
BEGIN
    INSERT OR IGNORE INTO rental_listings_history (
        rental_id, property_id,
        source_name, source_url, listing_ref,
        unit_number, monthly_rent, security_deposit,
        bedrooms, bathrooms, sqft,
        available_date, lease_term,
        status, days_on_market,
        listing_created_at, listing_expires_at,
        archive_reason
    ) VALUES (
        OLD.id, OLD.property_id,
        OLD.source_name, OLD.source_url, OLD.listing_ref,
        OLD.unit_number, OLD.monthly_rent, OLD.security_deposit,
        OLD.bedrooms, OLD.bathrooms, OLD.sqft,
        OLD.available_date, OLD.lease_term,
        OLD.status, OLD.days_on_market,
        OLD.created_at, OLD.expires_at,
        'deleted'
    );
END;
