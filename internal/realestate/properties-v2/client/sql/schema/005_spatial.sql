-- Spatial index using SQLite R*Tree module.
--
-- R*Tree requires integer primary keys, so we maintain a shadow table
-- (property_spatial_map) that bridges the TEXT property id to an integer rid.
-- Triggers keep everything in sync automatically.
--
-- Usage example (find all properties within a bounding box):
--   SELECT p.*
--   FROM property_rtree r
--   JOIN property_spatial_map m ON m.rid = r.id
--   JOIN properties p ON p.id = m.property_id
--   WHERE r.min_lat >= 33.5 AND r.max_lat <= 34.5
--     AND r.min_lng >= -118.5 AND r.max_lng <= -117.5;

-- Shadow table: text property_id <-> integer rid
CREATE TABLE IF NOT EXISTS property_spatial_map (
    rid         INTEGER PRIMARY KEY AUTOINCREMENT,
    property_id TEXT NOT NULL UNIQUE REFERENCES properties(id) ON DELETE CASCADE
);

-- R*Tree virtual table (point stored as min==max for lat/lng)
CREATE VIRTUAL TABLE IF NOT EXISTS property_rtree USING rtree(
    id,            -- integer (= property_spatial_map.rid)
    min_lat, max_lat,
    min_lng, max_lng
);

-- When a property is inserted: create spatial map entry and R*Tree row
CREATE TRIGGER IF NOT EXISTS trg_property_spatial_insert
AFTER INSERT ON properties
FOR EACH ROW
WHEN NEW.lat != 0 AND NEW.lng != 0
BEGIN
    INSERT INTO property_spatial_map(property_id) VALUES (NEW.id);
    INSERT INTO property_rtree(id, min_lat, max_lat, min_lng, max_lng)
        VALUES (last_insert_rowid(), NEW.lat, NEW.lat, NEW.lng, NEW.lng);
END;

-- When lat/lng changes: update R*Tree
CREATE TRIGGER IF NOT EXISTS trg_property_spatial_update
AFTER UPDATE OF lat, lng ON properties
FOR EACH ROW
WHEN NEW.lat != 0 AND NEW.lng != 0
BEGIN
    -- Upsert the spatial map entry
    INSERT OR IGNORE INTO property_spatial_map(property_id) VALUES (NEW.id);
    -- Update or insert R*Tree entry
    INSERT OR REPLACE INTO property_rtree(id, min_lat, max_lat, min_lng, max_lng)
        SELECT m.rid, NEW.lat, NEW.lat, NEW.lng, NEW.lng
        FROM property_spatial_map m
        WHERE m.property_id = NEW.id;
END;

-- When a property is deleted: cascade handled by FK + trigger
CREATE TRIGGER IF NOT EXISTS trg_property_spatial_delete
AFTER DELETE ON property_spatial_map
FOR EACH ROW
BEGIN
    DELETE FROM property_rtree WHERE id = OLD.rid;
END;


-- ── Rental listing spatial index ─────────────────────────────────────────────
-- Rentals pull coords from their parent property, but having a separate
-- rtree lets you filter by rent price AND bounding box in one pass.

CREATE TABLE IF NOT EXISTS rental_spatial_map (
    rid        INTEGER PRIMARY KEY AUTOINCREMENT,
    rental_id  TEXT NOT NULL UNIQUE REFERENCES rental_listings(id) ON DELETE CASCADE
);

CREATE VIRTUAL TABLE IF NOT EXISTS rental_rtree USING rtree(
    id,
    min_lat, max_lat,
    min_lng, max_lng
);

CREATE TRIGGER IF NOT EXISTS trg_rental_spatial_insert
AFTER INSERT ON rental_listings
FOR EACH ROW
BEGIN
    INSERT INTO rental_spatial_map(rental_id) VALUES (NEW.id);
    INSERT INTO rental_rtree(id, min_lat, max_lat, min_lng, max_lng)
        SELECT last_insert_rowid(), p.lat, p.lat, p.lng, p.lng
        FROM properties p WHERE p.id = NEW.property_id;
END;

CREATE TRIGGER IF NOT EXISTS trg_rental_spatial_delete
AFTER DELETE ON rental_spatial_map
FOR EACH ROW
BEGIN
    DELETE FROM rental_rtree WHERE id = OLD.rid;
END;


-- ── Sale listing spatial index ────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS listing_spatial_map (
    rid        INTEGER PRIMARY KEY AUTOINCREMENT,
    listing_id TEXT NOT NULL UNIQUE REFERENCES property_listings(id) ON DELETE CASCADE
);

CREATE VIRTUAL TABLE IF NOT EXISTS listing_rtree USING rtree(
    id,
    min_lat, max_lat,
    min_lng, max_lng
);

CREATE TRIGGER IF NOT EXISTS trg_listing_spatial_insert
AFTER INSERT ON property_listings
FOR EACH ROW
BEGIN
    INSERT INTO listing_spatial_map(listing_id) VALUES (NEW.id);
    INSERT INTO listing_rtree(id, min_lat, max_lat, min_lng, max_lng)
        SELECT last_insert_rowid(), p.lat, p.lat, p.lng, p.lng
        FROM properties p WHERE p.id = NEW.property_id;
END;

CREATE TRIGGER IF NOT EXISTS trg_listing_spatial_delete
AFTER DELETE ON listing_spatial_map
FOR EACH ROW
BEGIN
    DELETE FROM listing_rtree WHERE id = OLD.rid;
END;
