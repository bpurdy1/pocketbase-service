-- property_photos: links photos to a property.
-- source_url is where the photo was scraped/downloaded from.
-- local_path is optional if you store a local copy.

CREATE TABLE IF NOT EXISTS property_photos (
    id           TEXT PRIMARY KEY NOT NULL DEFAULT ('ph'||lower(hex(randomblob(7)))),
    property_id  TEXT NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    -- optional: tie a photo to a specific listing snapshot
    listing_id   TEXT REFERENCES property_listings(id) ON DELETE SET NULL,
    rental_id    TEXT REFERENCES rental_listings(id) ON DELETE SET NULL,
    -- photo location
    source_url   TEXT NOT NULL DEFAULT '',  -- original URL the photo was found at
    local_path   TEXT NOT NULL DEFAULT '',  -- local file path if cached
    -- metadata
    caption      TEXT NOT NULL DEFAULT '',
    is_primary   INTEGER NOT NULL DEFAULT 0, -- 1 = hero/cover photo
    width        INTEGER NOT NULL DEFAULT 0,
    height       INTEGER NOT NULL DEFAULT 0,
    size_bytes   INTEGER NOT NULL DEFAULT 0,
    mime_type    TEXT NOT NULL DEFAULT '',
    -- ordering
    sort_order   INTEGER NOT NULL DEFAULT 0,
    -- timestamps
    created_at   TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_photos_property
    ON property_photos(property_id);

CREATE INDEX IF NOT EXISTS idx_photos_listing
    ON property_photos(listing_id);

CREATE INDEX IF NOT EXISTS idx_photos_rental
    ON property_photos(rental_id);

CREATE INDEX IF NOT EXISTS idx_photos_primary
    ON property_photos(property_id, is_primary);


-- listing_sources: tracks every URL/site where a property was found.
-- Many-to-many between properties and external sources.

CREATE TABLE IF NOT EXISTS listing_sources (
    id           TEXT PRIMARY KEY NOT NULL DEFAULT ('s'||lower(hex(randomblob(7)))),
    property_id  TEXT NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    source_name  TEXT NOT NULL DEFAULT '',  -- 'Zillow','Redfin','MLS','LoopNet','CoStar'
    source_url   TEXT NOT NULL DEFAULT '',  -- full URL to the listing page
    source_type  TEXT NOT NULL DEFAULT '',  -- 'sale','rental','off_market','tax_record'
    last_seen_at TEXT NOT NULL DEFAULT (datetime('now')),
    first_seen_at TEXT NOT NULL DEFAULT (datetime('now')),
    is_active    INTEGER NOT NULL DEFAULT 1
);

-- One row per (property, source_url) combination
CREATE UNIQUE INDEX IF NOT EXISTS idx_sources_property_url
    ON listing_sources(property_id, source_url);

CREATE INDEX IF NOT EXISTS idx_sources_property
    ON listing_sources(property_id);

CREATE INDEX IF NOT EXISTS idx_sources_name
    ON listing_sources(source_name);

CREATE INDEX IF NOT EXISTS idx_sources_type
    ON listing_sources(source_type);

CREATE INDEX IF NOT EXISTS idx_sources_last_seen
    ON listing_sources(last_seen_at);
