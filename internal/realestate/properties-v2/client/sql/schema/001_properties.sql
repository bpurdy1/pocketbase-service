-- Properties: canonical record for a physical address.
-- One row per unique address. Listings and rentals reference this table.

CREATE TABLE IF NOT EXISTS properties (
    id          TEXT PRIMARY KEY NOT NULL DEFAULT ('p'||lower(hex(randomblob(7)))),
    -- address
    address     TEXT NOT NULL DEFAULT '',
    city        TEXT NOT NULL DEFAULT '',
    state       TEXT NOT NULL DEFAULT '',
    zip_code    TEXT NOT NULL DEFAULT '',
    county      TEXT NOT NULL DEFAULT '',
    -- geo coords (also indexed via R*Tree in 005_spatial.sql)
    lat         REAL NOT NULL DEFAULT 0,
    lng         REAL NOT NULL DEFAULT 0,
    -- property details
    property_name    TEXT NOT NULL DEFAULT '',
    property_type    TEXT NOT NULL DEFAULT '',  -- 'sfr','mfr','condo','land','commercial'
    bedrooms         REAL NOT NULL DEFAULT 0,
    bathrooms        REAL NOT NULL DEFAULT 0,
    sqft             REAL NOT NULL DEFAULT 0,
    building_sf      REAL NOT NULL DEFAULT 0,
    lot_sf           REAL NOT NULL DEFAULT 0,
    year_built       INTEGER NOT NULL DEFAULT 0,
    number_of_units  INTEGER NOT NULL DEFAULT 0,
    organization     TEXT NOT NULL DEFAULT '',
    notes            TEXT NOT NULL DEFAULT '',
    -- timestamps (auto-managed)
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Enforce uniqueness on address so one property row per physical location
CREATE UNIQUE INDEX IF NOT EXISTS idx_properties_address
    ON properties(address);

CREATE INDEX IF NOT EXISTS idx_properties_city_state
    ON properties(city, state);

CREATE INDEX IF NOT EXISTS idx_properties_type
    ON properties(property_type);

CREATE INDEX IF NOT EXISTS idx_properties_org
    ON properties(organization);

CREATE INDEX IF NOT EXISTS idx_properties_created
    ON properties(created_at);

-- Auto-update updated_at on any change
CREATE TRIGGER IF NOT EXISTS trg_properties_updated_at
AFTER UPDATE ON properties
FOR EACH ROW
WHEN NEW.updated_at = OLD.updated_at
BEGIN
    UPDATE properties SET updated_at = datetime('now') WHERE id = NEW.id;
END;
