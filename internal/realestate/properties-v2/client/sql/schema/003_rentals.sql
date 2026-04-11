-- rental_listings: one active rental listing per property address.
-- Same TTL pattern as property_listings.

CREATE TABLE IF NOT EXISTS rental_listings (
    id           TEXT PRIMARY KEY NOT NULL DEFAULT ('r'||lower(hex(randomblob(7)))),
    property_id  TEXT NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    -- where the listing was found
    source_name  TEXT NOT NULL DEFAULT '',   -- e.g. 'Zillow','Apartments.com','Craigslist'
    source_url   TEXT NOT NULL DEFAULT '',
    listing_ref  TEXT NOT NULL DEFAULT '',   -- external reference/ID from source
    -- rental pricing
    monthly_rent     REAL NOT NULL DEFAULT 0,
    security_deposit REAL NOT NULL DEFAULT 0,
    rent_per_sqft    REAL NOT NULL DEFAULT 0,
    -- unit details (may differ from base property if multi-unit)
    unit_number  TEXT NOT NULL DEFAULT '',
    bedrooms     REAL NOT NULL DEFAULT 0,
    bathrooms    REAL NOT NULL DEFAULT 0,
    sqft         REAL NOT NULL DEFAULT 0,
    -- listing details
    available_date   TEXT NOT NULL DEFAULT '',
    lease_term       TEXT NOT NULL DEFAULT '', -- 'month-to-month','6mo','12mo','24mo'
    pets_allowed     INTEGER NOT NULL DEFAULT 0, -- 0=no,1=yes
    furnished        INTEGER NOT NULL DEFAULT 0,
    status           TEXT NOT NULL DEFAULT 'active', -- 'active','rented','off_market'
    days_on_market   INTEGER NOT NULL DEFAULT 0,
    -- TTL
    expires_at   TEXT NOT NULL DEFAULT (datetime('now', '+30 days')),
    -- timestamps
    created_at   TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at   TEXT NOT NULL DEFAULT (datetime('now'))
);

-- One active rental per property per unit
CREATE UNIQUE INDEX IF NOT EXISTS idx_rental_listings_property_unit
    ON rental_listings(property_id, unit_number);

CREATE INDEX IF NOT EXISTS idx_rental_listings_source
    ON rental_listings(source_name);

CREATE INDEX IF NOT EXISTS idx_rental_listings_status
    ON rental_listings(status);

CREATE INDEX IF NOT EXISTS idx_rental_listings_expires
    ON rental_listings(expires_at);

CREATE INDEX IF NOT EXISTS idx_rental_listings_rent
    ON rental_listings(monthly_rent);

CREATE INDEX IF NOT EXISTS idx_rental_listings_available
    ON rental_listings(available_date);

CREATE TRIGGER IF NOT EXISTS trg_rental_listings_updated_at
AFTER UPDATE ON rental_listings
FOR EACH ROW
WHEN NEW.updated_at = OLD.updated_at
BEGIN
    UPDATE rental_listings SET updated_at = datetime('now') WHERE id = NEW.id;
END;


-- rental_listings_history: immutable archive of past rental listings.

CREATE TABLE IF NOT EXISTS rental_listings_history (
    id               TEXT PRIMARY KEY NOT NULL DEFAULT ('rh'||lower(hex(randomblob(7)))),
    rental_id        TEXT NOT NULL,  -- original rental_listings.id
    property_id      TEXT NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    source_name      TEXT NOT NULL DEFAULT '',
    source_url       TEXT NOT NULL DEFAULT '',
    listing_ref      TEXT NOT NULL DEFAULT '',
    unit_number      TEXT NOT NULL DEFAULT '',
    monthly_rent     REAL NOT NULL DEFAULT 0,
    security_deposit REAL NOT NULL DEFAULT 0,
    bedrooms         REAL NOT NULL DEFAULT 0,
    bathrooms        REAL NOT NULL DEFAULT 0,
    sqft             REAL NOT NULL DEFAULT 0,
    available_date   TEXT NOT NULL DEFAULT '',
    lease_term       TEXT NOT NULL DEFAULT '',
    status           TEXT NOT NULL DEFAULT '',
    days_on_market   INTEGER NOT NULL DEFAULT 0,
    -- when active
    listing_created_at TEXT NOT NULL DEFAULT '',
    listing_expires_at TEXT NOT NULL DEFAULT '',
    archive_reason   TEXT NOT NULL DEFAULT '', -- 'expired','rented','superseded','removed'
    archived_at      TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_rlh_property
    ON rental_listings_history(property_id);

CREATE INDEX IF NOT EXISTS idx_rlh_rental
    ON rental_listings_history(rental_id);

CREATE INDEX IF NOT EXISTS idx_rlh_archived
    ON rental_listings_history(archived_at);

CREATE INDEX IF NOT EXISTS idx_rlh_status
    ON rental_listings_history(status);
