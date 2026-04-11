-- property_listings: one active for-sale listing per property address.
-- Expired rows are moved to property_listings_history by the application.
-- TTL is stored as expires_at; bump it with UpdateListingTTL query.

CREATE TABLE IF NOT EXISTS property_listings (
    id           TEXT PRIMARY KEY NOT NULL DEFAULT ('l'||lower(hex(randomblob(7)))),
    property_id  TEXT NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    -- where the listing was found
    source_name  TEXT NOT NULL DEFAULT '',   -- e.g. 'Zillow','Redfin','MLS','LoopNet'
    source_url   TEXT NOT NULL DEFAULT '',   -- deep link to the listing page
    mls_id       TEXT NOT NULL DEFAULT '',   -- MLS number if available
    -- pricing
    list_price   REAL NOT NULL DEFAULT 0,
    price_per_sqft REAL NOT NULL DEFAULT 0,
    -- listing status
    status           TEXT NOT NULL DEFAULT 'active', -- 'active','pending','sold','off_market'
    days_on_market   INTEGER NOT NULL DEFAULT 0,
    -- TTL: listings auto-expire; bump with UpdateListingTTL
    expires_at   TEXT NOT NULL DEFAULT (datetime('now', '+30 days')),
    -- timestamps (auto-managed)
    created_at   TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at   TEXT NOT NULL DEFAULT (datetime('now'))
);

-- One active listing per property (unique per address via property_id)
CREATE UNIQUE INDEX IF NOT EXISTS idx_property_listings_property
    ON property_listings(property_id);

CREATE INDEX IF NOT EXISTS idx_property_listings_source
    ON property_listings(source_name);

CREATE INDEX IF NOT EXISTS idx_property_listings_status
    ON property_listings(status);

CREATE INDEX IF NOT EXISTS idx_property_listings_expires
    ON property_listings(expires_at);

CREATE INDEX IF NOT EXISTS idx_property_listings_price
    ON property_listings(list_price);

CREATE TRIGGER IF NOT EXISTS trg_property_listings_updated_at
AFTER UPDATE ON property_listings
FOR EACH ROW
WHEN NEW.updated_at = OLD.updated_at
BEGIN
    UPDATE property_listings SET updated_at = datetime('now') WHERE id = NEW.id;
END;


-- property_listings_history: immutable archive of all past sale listings.
-- Populated by the application when a listing expires or is superseded.

CREATE TABLE IF NOT EXISTS property_listings_history (
    id           TEXT PRIMARY KEY NOT NULL DEFAULT ('lh'||lower(hex(randomblob(7)))),
    -- original listing data (denormalized for historical integrity)
    listing_id   TEXT NOT NULL,  -- original property_listings.id
    property_id  TEXT NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    source_name  TEXT NOT NULL DEFAULT '',
    source_url   TEXT NOT NULL DEFAULT '',
    mls_id       TEXT NOT NULL DEFAULT '',
    list_price   REAL NOT NULL DEFAULT 0,
    price_per_sqft REAL NOT NULL DEFAULT 0,
    status       TEXT NOT NULL DEFAULT '',
    days_on_market INTEGER NOT NULL DEFAULT 0,
    -- when the listing was active
    listing_created_at TEXT NOT NULL DEFAULT '',
    listing_expires_at TEXT NOT NULL DEFAULT '',
    -- reason this record was archived
    archive_reason TEXT NOT NULL DEFAULT '', -- 'expired','sold','superseded','removed'
    archived_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_plh_property
    ON property_listings_history(property_id);

CREATE INDEX IF NOT EXISTS idx_plh_listing
    ON property_listings_history(listing_id);

CREATE INDEX IF NOT EXISTS idx_plh_archived
    ON property_listings_history(archived_at);

CREATE INDEX IF NOT EXISTS idx_plh_status
    ON property_listings_history(status);
