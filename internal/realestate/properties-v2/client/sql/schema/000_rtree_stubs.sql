-- sqlc stubs for R*Tree virtual tables and their shadow maps.
-- These are plain CREATE TABLE definitions so sqlc can infer column types.
-- The real runtime schema (005_spatial.sql) uses CREATE VIRTUAL TABLE USING rtree
-- and is excluded from sqlc processing.

CREATE TABLE IF NOT EXISTS property_spatial_map (
    rid         INTEGER PRIMARY KEY AUTOINCREMENT,
    property_id TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS property_rtree (
    id      INTEGER PRIMARY KEY,
    min_lat REAL NOT NULL,
    max_lat REAL NOT NULL,
    min_lng REAL NOT NULL,
    max_lng REAL NOT NULL
);

CREATE TABLE IF NOT EXISTS listing_spatial_map (
    rid        INTEGER PRIMARY KEY AUTOINCREMENT,
    listing_id TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS listing_rtree (
    id      INTEGER PRIMARY KEY,
    min_lat REAL NOT NULL,
    max_lat REAL NOT NULL,
    min_lng REAL NOT NULL,
    max_lng REAL NOT NULL
);

CREATE TABLE IF NOT EXISTS rental_spatial_map (
    rid       INTEGER PRIMARY KEY AUTOINCREMENT,
    rental_id TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS rental_rtree (
    id      INTEGER PRIMARY KEY,
    min_lat REAL NOT NULL,
    max_lat REAL NOT NULL,
    min_lng REAL NOT NULL,
    max_lng REAL NOT NULL
);
