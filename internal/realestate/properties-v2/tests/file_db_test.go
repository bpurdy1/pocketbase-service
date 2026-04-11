package tests

// file_db_test.go — full end-to-end tests against a real SQLite file.
//
// These tests use the fileDB client (created in TestMain) which opens a fresh
// temp file, runs all migrations (including 006_history_triggers), then deletes
// the file on cleanup. This catches issues that only appear with a real file
// (WAL, trigger persistence, migration tracking) and not in :memory: DBs.
//
// Design rules for this file:
//   - All tests share fileDB — no per-test teardown.
//   - Every seedProperty call must use a unique address (address has a UNIQUE constraint).
//   - Non-spatial tests must use coords (0, 0) so they don't pollute spatial queries.

import (
	"testing"

	"sqlite-realestate/client"
)

// ── Migration ─────────────────────────────────────────────────────────────────

func TestFileDB_MigrationsApplied(t *testing.T) { // schema_migrations table should have one row per migration file
	var count int
	if err := fileDB.RawDB().QueryRowContext(ctx,
		`SELECT COUNT(*) FROM schema_migrations`,
	).Scan(&count); err != nil {
		t.Fatalf("query schema_migrations: %v", err)
	}
	if count != 6 {
		t.Errorf("migrations applied = %d, want 6", count)
	}
}

func TestFileDB_MigrationsIdempotent(t *testing.T) {
	// Running Migrate again on the same file should be a no-op
	if err := client.Migrate(ctx, fileDB.RawDB()); err != nil {
		t.Fatalf("second Migrate call: %v", err)
	}
	var count int
	if err := fileDB.RawDB().QueryRowContext(ctx,
		`SELECT COUNT(*) FROM schema_migrations`,
	).Scan(&count); err != nil {
		t.Fatalf("query schema_migrations: %v", err)
	}
	if count != 6 {
		t.Errorf("after re-migrate count = %d, want 6 (no duplicates)", count)
	}
}

// ── Properties ───────────────────────────────────────────────────────────────

func TestFileDB_Property_CreateAndRead(t *testing.T) {
	prop := seedProperty(t, fileDB, func(p *client.CreatePropertyParams) {
		p.Address = "filedb-prop-create"
		p.Lat, p.Lng = 0, 0
	})
	got, err := fileDB.GetProperty(ctx, prop.ID)
	if err != nil {
		t.Fatalf("GetProperty: %v", err)
	}
	if got.ID != prop.ID {
		t.Errorf("ID = %q, want %q", got.ID, prop.ID)
	}
	if got.Address != prop.Address {
		t.Errorf("Address = %q, want %q", got.Address, prop.Address)
	}
}

func TestFileDB_Property_UpdateWritesHistory(t *testing.T) {
	prop := seedProperty(t, fileDB, func(p *client.CreatePropertyParams) {
		p.Address = "filedb-prop-update"
		p.Lat, p.Lng = 0, 0
	})
	updatePropertyCity(t, fileDB, prop, "San Francisco")

	history, err := fileDB.GetPropertyHistory(ctx, client.GetPropertyHistoryParams{PropertyID: prop.ID, Limit: 10})
	if err != nil {
		t.Fatalf("GetPropertyHistory: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("history len = %d, want 1", len(history))
	}
	if history[0].ChangeType != "updated" {
		t.Errorf("ChangeType = %q, want updated", history[0].ChangeType)
	}
	if history[0].City != "Los Angeles" {
		t.Errorf("snapshot City = %q, want Los Angeles (old value)", history[0].City)
	}
}

func TestFileDB_Property_DeleteWritesHistory(t *testing.T) {
	prop := seedProperty(t, fileDB, func(p *client.CreatePropertyParams) {
		p.Address = "filedb-prop-delete"
		p.Lat, p.Lng = 0, 0
	})
	if err := fileDB.DeleteProperty(ctx, prop.ID); err != nil {
		t.Fatalf("DeleteProperty: %v", err)
	}

	history, err := fileDB.GetPropertyHistory(ctx, client.GetPropertyHistoryParams{PropertyID: prop.ID, Limit: 10})
	if err != nil {
		t.Fatalf("GetPropertyHistory: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("history len = %d, want 1", len(history))
	}
	if history[0].ChangeType != "deleted" {
		t.Errorf("ChangeType = %q, want deleted", history[0].ChangeType)
	}
}

// ── Listings ─────────────────────────────────────────────────────────────────

func TestFileDB_Listing_CreateAndRead(t *testing.T) {
	prop := seedProperty(t, fileDB, func(p *client.CreatePropertyParams) {
		p.Address = "filedb-listing-create"
		p.Lat, p.Lng = 0, 0
	})
	listing := seedListing(t, fileDB, prop.ID)

	got, err := fileDB.GetListing(ctx, listing.ID)
	if err != nil {
		t.Fatalf("GetListing: %v", err)
	}
	if got.ID != listing.ID {
		t.Errorf("ID = %q, want %q", got.ID, listing.ID)
	}
}

func TestFileDB_Listing_UpdateWritesHistory(t *testing.T) {
	prop := seedProperty(t, fileDB, func(p *client.CreatePropertyParams) {
		p.Address = "filedb-listing-update"
		p.Lat, p.Lng = 0, 0
	})
	listing := seedListing(t, fileDB, prop.ID)

	if _, err := fileDB.UpdateListingStatus(ctx, client.UpdateListingStatusParams{Status: "pending", ID: listing.ID}); err != nil {
		t.Fatalf("UpdateListingStatus: %v", err)
	}

	history, err := fileDB.GetListingHistory(ctx, client.GetListingHistoryParams{PropertyID: prop.ID, Limit: 10})
	if err != nil {
		t.Fatalf("GetListingHistory: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("history len = %d, want 1", len(history))
	}
	if history[0].Status != "active" {
		t.Errorf("snapshot Status = %q, want active (old value)", history[0].Status)
	}
}

func TestFileDB_Listing_DeleteAutoArchives(t *testing.T) {
	prop := seedProperty(t, fileDB, func(p *client.CreatePropertyParams) {
		p.Address = "filedb-listing-delete"
		p.Lat, p.Lng = 0, 0
	})
	listing := seedListing(t, fileDB, prop.ID)

	if err := fileDB.DeleteListing(ctx, listing.ID); err != nil {
		t.Fatalf("DeleteListing: %v", err)
	}

	history, err := fileDB.GetListingHistoryByListingID(ctx, listing.ID)
	if err != nil {
		t.Fatalf("GetListingHistoryByListingID: %v", err)
	}
	if len(history) == 0 {
		t.Fatal("expected history row after delete, got none")
	}
	if history[0].ArchiveReason != "deleted" {
		t.Errorf("ArchiveReason = %q, want deleted", history[0].ArchiveReason)
	}
}

func TestFileDB_Listing_TTLBump(t *testing.T) {
	prop := seedProperty(t, fileDB, func(p *client.CreatePropertyParams) {
		p.Address = "filedb-listing-ttl"
		p.Lat, p.Lng = 0, 0
	})
	listing := seedListing(t, fileDB, prop.ID)
	newExpiry := futureDate(60)

	updated, err := fileDB.UpdateListingTTL(ctx, client.UpdateListingTTLParams{ExpiresAt: newExpiry, ID: listing.ID})
	if err != nil {
		t.Fatalf("UpdateListingTTL: %v", err)
	}
	if updated.ExpiresAt != newExpiry {
		t.Errorf("ExpiresAt = %q, want %q", updated.ExpiresAt, newExpiry)
	}

	// TTL bump is an UPDATE on expires_at → history trigger fires
	history, err := fileDB.GetListingHistory(ctx, client.GetListingHistoryParams{PropertyID: prop.ID, Limit: 10})
	if err != nil {
		t.Fatalf("GetListingHistory: %v", err)
	}
	if len(history) != 1 {
		t.Errorf("history len = %d, want 1 after TTL bump", len(history))
	}
}

// ── Rentals ───────────────────────────────────────────────────────────────────

func TestFileDB_Rental_CreateAndRead(t *testing.T) {
	prop := seedProperty(t, fileDB, func(p *client.CreatePropertyParams) {
		p.Address = "filedb-rental-create"
		p.Lat, p.Lng = 0, 0
	})
	rental := seedRental(t, fileDB, prop.ID)

	got, err := fileDB.GetRental(ctx, rental.ID)
	if err != nil {
		t.Fatalf("GetRental: %v", err)
	}
	if got.ID != rental.ID {
		t.Errorf("ID = %q, want %q", got.ID, rental.ID)
	}
}

func TestFileDB_Rental_UpdateWritesHistory(t *testing.T) {
	prop := seedProperty(t, fileDB, func(p *client.CreatePropertyParams) {
		p.Address = "filedb-rental-update"
		p.Lat, p.Lng = 0, 0
	})
	rental := seedRental(t, fileDB, prop.ID)
	updateRentalRent(t, fileDB, rental, 3000)

	history, err := fileDB.GetRentalHistory(ctx, client.GetRentalHistoryParams{PropertyID: prop.ID, Limit: 10})
	if err != nil {
		t.Fatalf("GetRentalHistory: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("history len = %d, want 1", len(history))
	}
	if history[0].MonthlyRent != 2000 {
		t.Errorf("snapshot MonthlyRent = %v, want 2000 (old value)", history[0].MonthlyRent)
	}
}

func TestFileDB_Rental_DeleteAutoArchives(t *testing.T) {
	prop := seedProperty(t, fileDB, func(p *client.CreatePropertyParams) {
		p.Address = "filedb-rental-delete"
		p.Lat, p.Lng = 0, 0
	})
	rental := seedRental(t, fileDB, prop.ID)

	if err := fileDB.DeleteRental(ctx, rental.ID); err != nil {
		t.Fatalf("DeleteRental: %v", err)
	}

	history, err := fileDB.GetRentalHistoryByRentalID(ctx, rental.ID)
	if err != nil {
		t.Fatalf("GetRentalHistoryByRentalID: %v", err)
	}
	if len(history) == 0 {
		t.Fatal("expected history row after delete, got none")
	}
	if history[0].ArchiveReason != "deleted" {
		t.Errorf("ArchiveReason = %q, want deleted", history[0].ArchiveReason)
	}
}

// ── Photos ────────────────────────────────────────────────────────────────────

func TestFileDB_Photo_AddAndList(t *testing.T) {
	prop := seedProperty(t, fileDB, func(p *client.CreatePropertyParams) {
		p.Address = "filedb-photo"
		p.Lat, p.Lng = 0, 0
	})
	seedPhoto(t, fileDB, prop.ID, func(p *client.AddPhotoParams) {
		p.SourceUrl = "https://example.com/house.jpg"
		p.IsPrimary = 1
	})
	seedPhoto(t, fileDB, prop.ID, func(p *client.AddPhotoParams) {
		p.SourceUrl = "https://example.com/kitchen.jpg"
	})

	photos, err := fileDB.ListPhotosForProperty(ctx, prop.ID)
	if err != nil {
		t.Fatalf("ListPhotosForProperty: %v", err)
	}
	if len(photos) != 2 {
		t.Errorf("len = %d, want 2", len(photos))
	}
	if photos[0].IsPrimary != 1 {
		t.Error("primary photo should be first (sorted by is_primary DESC)")
	}
}

// ── Listing sources ───────────────────────────────────────────────────────────

func TestFileDB_ListingSource_UpsertIdempotent(t *testing.T) {
	prop := seedProperty(t, fileDB, func(p *client.CreatePropertyParams) {
		p.Address = "filedb-listing-source"
		p.Lat, p.Lng = 0, 0
	})
	params := client.UpsertListingSourceParams{PropertyID: prop.ID,
		SourceName:  "Zillow",
		SourceUrl:   "https://zillow.com/test",
		SourceType:  "sale",
		LastSeenAt:  "2026-04-01T00:00:00Z",
		FirstSeenAt: "2026-04-01T00:00:00Z",
	}

	if _, err := fileDB.UpsertListingSource(ctx, params); err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	params.LastSeenAt = "2026-04-04T00:00:00Z"
	if _, err := fileDB.UpsertListingSource(ctx, params); err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	sources, err := fileDB.GetListingSources(ctx, prop.ID)
	if err != nil {
		t.Fatalf("GetListingSources: %v", err)
	}
	if len(sources) != 1 {
		t.Errorf("len = %d, want 1 (upsert must not duplicate)", len(sources))
	}
	if sources[0].LastSeenAt != "2026-04-04T00:00:00Z" {
		t.Errorf("LastSeenAt = %q, want updated value", sources[0].LastSeenAt)
	}
}

// ── Spatial ───────────────────────────────────────────────────────────────────

// TestFileDB_Spatial_TriggerAndQuery uses NYC-area coords (far from the LA
// coords used by other tests and FullLifecycle) so spatial results are isolated.
func TestFileDB_Spatial_TriggerAndQuery(t *testing.T) {
	// NYC downtown — query center
	const nycLat, nycLng = 40.712, -74.006
	// Insert populates R*Tree via trigger
	seedPropertyAt(t, fileDB, "filedb-spatial-near", nycLat, nycLng)
	seedPropertyAt(t, fileDB, "filedb-spatial-far", 32.714, -117.173) // San Diego — ~3900km away

	results, err := fileDB.FindPropertiesNearby(ctx, nycLat, nycLng, 50)
	if err != nil {
		t.Fatalf("FindPropertiesNearby: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("len = %d, want 1 (only the NYC property within 50km)", len(results))
	}
	if results[0].Address != "filedb-spatial-near" {
		t.Errorf("Address = %q, want filedb-spatial-near", results[0].Address)
	}
}

// ── Full lifecycle ────────────────────────────────────────────────────────────

func TestFileDB_FullLifecycle(t *testing.T) {
	// 1. Create property — unique LA-area coords not used elsewhere in file_db tests
	prop := seedProperty(t, fileDB, func(p *client.CreatePropertyParams) {
		p.Address = "99 Lifecycle Ave"
		p.Lat = 34.052
		p.Lng = -118.243
	})

	// 2. Add a sale listing
	listing := seedListing(t, fileDB, prop.ID, func(p *client.CreateListingParams) {
		p.ListPrice = 750000
	})

	// 3. Add a rental on the same property
	rental := seedRental(t, fileDB, prop.ID, func(p *client.CreateRentalParams) {
		p.MonthlyRent = 3200
	})

	// 4. Add photos
	seedPhoto(t, fileDB, prop.ID, func(p *client.AddPhotoParams) {
		p.SourceUrl = "https://example.com/front.jpg"
		p.IsPrimary = 1
	})

	// 5. Update listing price — history trigger fires
	if _, err := fileDB.UpdateListing(ctx, client.UpdateListingParams{ID: listing.ID,
		SourceName:   listing.SourceName,
		SourceUrl:    listing.SourceUrl,
		MlsID:        listing.MlsID,
		ListPrice:    800000,
		PricePerSqft: 533.33,
		Status:       listing.Status,
		DaysOnMarket: listing.DaysOnMarket,
		ExpiresAt:    listing.ExpiresAt,
	}); err != nil {
		t.Fatalf("UpdateListing: %v", err)
	}

	// 6. Rent goes up — history trigger fires
	updateRentalRent(t, fileDB, rental, 3500)

	// 7. Property address corrected — history trigger fires
	updatePropertyCity(t, fileDB, prop, "West Hollywood")

	// 8. Verify all history recorded
	propHistory, err := fileDB.GetPropertyHistory(ctx, client.GetPropertyHistoryParams{PropertyID: prop.ID, Limit: 10})
	if err != nil {
		t.Fatalf("property history: %v", err)
	}
	if len(propHistory) != 1 {
		t.Errorf("property history len = %d, want 1", len(propHistory))
	}

	listingHistory, err := fileDB.GetListingHistory(ctx, client.GetListingHistoryParams{PropertyID: prop.ID, Limit: 10})
	if err != nil {
		t.Fatalf("listing history: %v", err)
	}
	if len(listingHistory) != 1 {
		t.Errorf("listing history len = %d, want 1", len(listingHistory))
	}
	if listingHistory[0].ListPrice != 750000 {
		t.Errorf("listing snapshot ListPrice = %v, want 750000", listingHistory[0].ListPrice)
	}

	rentalHistory, err := fileDB.GetRentalHistory(ctx, client.GetRentalHistoryParams{PropertyID: prop.ID, Limit: 10})
	if err != nil {
		t.Fatalf("rental history: %v", err)
	}
	if len(rentalHistory) != 1 {
		t.Errorf("rental history len = %d, want 1", len(rentalHistory))
	}
	if rentalHistory[0].MonthlyRent != 3200 {
		t.Errorf("rental snapshot MonthlyRent = %v, want 3200", rentalHistory[0].MonthlyRent)
	}

	// 9. Delete listing — BEFORE DELETE trigger archives it
	if err := fileDB.DeleteListing(ctx, listing.ID); err != nil {
		t.Fatalf("DeleteListing: %v", err)
	}
	allListingHistory, err := fileDB.GetListingHistoryByListingID(ctx, listing.ID)
	if err != nil {
		t.Fatalf("all listing history: %v", err)
	}
	// 1 update snapshot + 1 delete snapshot
	if len(allListingHistory) != 2 {
		t.Errorf("all listing history len = %d, want 2", len(allListingHistory))
	}

	// 10. Spatial query: only "99 Lifecycle Ave" is at LA coords among file_db tests
	results, err := fileDB.FindPropertiesNearby(ctx, 34.052, -118.243, 10)
	if err != nil {
		t.Fatalf("FindPropertiesNearby: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("spatial results = %d, want 1", len(results))
	}
}
