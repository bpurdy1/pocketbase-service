package tests

import (
	"testing"

	"sqlite-realestate/client"
)

// Verifies that rental_listings_history receives a snapshot
// whenever a rental_listing row is updated or deleted.
//
// Trigger source: 006_history_triggers.sql
//   - trg_rental_history_update  → fires AFTER UPDATE on rental_listings
//   - trg_rental_history_delete  → fires BEFORE DELETE on rental_listings

// helper: bump monthly rent on a rental and return the updated row
func updateRentalRent(t *testing.T, c *client.Client, r client.RentalListing, rent float64) client.RentalListing {
	t.Helper()
	updated, err := c.UpdateRental(ctx, client.UpdateRentalParams{
		ID:              r.ID,
		SourceName:      r.SourceName,
		SourceUrl:       r.SourceUrl,
		ListingRef:      r.ListingRef,
		MonthlyRent:     rent,
		SecurityDeposit: r.SecurityDeposit,
		RentPerSqft:     r.RentPerSqft,
		UnitNumber:      r.UnitNumber,
		Bedrooms:        r.Bedrooms,
		Bathrooms:       r.Bathrooms,
		Sqft:            r.Sqft,
		AvailableDate:   r.AvailableDate,
		LeaseTerm:       r.LeaseTerm,
		PetsAllowed:     r.PetsAllowed,
		Furnished:       r.Furnished,
		Status:          r.Status,
		DaysOnMarket:    r.DaysOnMarket,
		ExpiresAt:       r.ExpiresAt,
	})
	if err != nil {
		t.Fatalf("UpdateRental: %v", err)
	}
	return client.WrapRental(updated)
}

// -- INSERT -------------------------------------------------------------------

func TestRentalHistory_NoRowOnInsert(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	rental := seedRental(t, c, prop.ID)

	history, err := c.GetRentalHistory(ctx, client.GetRentalHistoryParams{
		PropertyID: prop.ID,
		Limit:      10,
		Offset:     0,
	})
	if err != nil {
		t.Fatalf("GetRentalHistory: %v", err)
	}
	if len(history) != 0 {
		t.Errorf("len = %d, want 0 — insert should not write history (rental id: %s)", len(history), rental.ID)
	}
}

// -- UPDATE -------------------------------------------------------------------

func TestRentalHistory_SingleUpdate(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	rental := seedRental(t, c, prop.ID)
	updateRentalRent(t, c, rental, 2500)

	history, err := c.GetRentalHistory(ctx, client.GetRentalHistoryParams{
		PropertyID: prop.ID,
		Limit:      10,
		Offset:     0,
	})
	if err != nil {
		t.Fatalf("GetRentalHistory: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("len = %d, want 1", len(history))
	}

	row := history[0]
	if row.ArchiveReason != "updated" {
		t.Errorf("ArchiveReason = %q, want updated", row.ArchiveReason)
	}
	// Snapshot holds OLD rent
	if row.MonthlyRent != 2000 {
		t.Errorf("snapshot MonthlyRent = %v, want 2000 (pre-update value)", row.MonthlyRent)
	}
	if row.RentalID != rental.ID {
		t.Errorf("RentalID = %q, want %q", row.RentalID, rental.ID)
	}
}

func TestRentalHistory_MultipleUpdates_OneRowEach(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	rental := seedRental(t, c, prop.ID)

	for _, rent := range []float64{2100, 2200, 2300} {
		rental = updateRentalRent(t, c, rental, rent)
	}

	history, err := c.GetRentalHistory(ctx, client.GetRentalHistoryParams{
		PropertyID: prop.ID,
		Limit:      10,
		Offset:     0,
	})
	if err != nil {
		t.Fatalf("GetRentalHistory: %v", err)
	}
	if len(history) != 3 {
		t.Errorf("len = %d, want 3 (one snapshot per update)", len(history))
	}
}

func TestRentalHistory_SnapshotPreservesAllFields(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	rental := seedRental(t, c, prop.ID)
	updateRentalRent(t, c, rental, 2500)

	history, err := c.GetRentalHistory(ctx, client.GetRentalHistoryParams{
		PropertyID: prop.ID,
		Limit:      1,
		Offset:     0,
	})
	if err != nil {
		t.Fatalf("GetRentalHistory: %v", err)
	}
	if len(history) == 0 {
		t.Fatal("no history rows")
	}

	row := history[0]
	if row.SourceName != rental.SourceName {
		t.Errorf("SourceName = %q, want %q", row.SourceName, rental.SourceName)
	}
	if row.Bedrooms != rental.Bedrooms {
		t.Errorf("Bedrooms = %v, want %v", row.Bedrooms, rental.Bedrooms)
	}
	if row.LeaseTerm != rental.LeaseTerm {
		t.Errorf("LeaseTerm = %q, want %q", row.LeaseTerm, rental.LeaseTerm)
	}
	if row.AvailableDate != rental.AvailableDate {
		t.Errorf("AvailableDate = %q, want %q", row.AvailableDate, rental.AvailableDate)
	}
}

func TestRentalHistory_StatusChange(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	rental := seedRental(t, c, prop.ID)

	if _, err := c.UpdateRentalStatus(ctx, client.UpdateRentalStatusParams{
		Status: "rented",
		ID:     rental.ID,
	}); err != nil {
		t.Fatalf("UpdateRentalStatus: %v", err)
	}

	history, err := c.GetRentalHistory(ctx, client.GetRentalHistoryParams{
		PropertyID: prop.ID,
		Limit:      10,
		Offset:     0,
	})
	if err != nil {
		t.Fatalf("GetRentalHistory: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("len = %d, want 1", len(history))
	}
	// Snapshot holds OLD status
	if history[0].Status != "active" {
		t.Errorf("snapshot Status = %q, want active (pre-update value)", history[0].Status)
	}
}

func TestRentalHistory_TTLBump(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	rental := seedRental(t, c, prop.ID)

	if _, err := c.UpdateRentalTTL(ctx, client.UpdateRentalTTLParams{
		ExpiresAt: futureDate(60),
		ID:        rental.ID,
	}); err != nil {
		t.Fatalf("UpdateRentalTTL: %v", err)
	}

	history, err := c.GetRentalHistory(ctx, client.GetRentalHistoryParams{
		PropertyID: prop.ID,
		Limit:      10,
		Offset:     0,
	})
	if err != nil {
		t.Fatalf("GetRentalHistory: %v", err)
	}
	// TTL bump is an UPDATE so trigger fires
	if len(history) != 1 {
		t.Errorf("len = %d, want 1 (TTL bump is an update)", len(history))
	}
	// Snapshot holds OLD expiry
	if history[0].ListingExpiresAt == futureDate(60) {
		t.Error("snapshot should hold old ExpiresAt, not the new value")
	}
}

// -- DELETE -------------------------------------------------------------------

func TestRentalHistory_DeleteAutoArchives(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	rental := seedRental(t, c, prop.ID)

	// Delete without manually archiving — BEFORE DELETE trigger auto-archives
	if err := c.DeleteRental(ctx, rental.ID); err != nil {
		t.Fatalf("DeleteRental: %v", err)
	}

	history, err := c.GetRentalHistoryByRentalID(ctx, rental.ID)
	if err != nil {
		t.Fatalf("GetRentalHistoryByRentalID: %v", err)
	}
	if len(history) == 0 {
		t.Fatal("expected history row after delete, got none")
	}
	if history[0].ArchiveReason != "deleted" {
		t.Errorf("ArchiveReason = %q, want deleted", history[0].ArchiveReason)
	}
	if history[0].MonthlyRent != rental.MonthlyRent {
		t.Errorf("snapshot MonthlyRent = %v, want %v", history[0].MonthlyRent, rental.MonthlyRent)
	}
}

func TestRentalHistory_UpdateThenDelete(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	rental := seedRental(t, c, prop.ID)

	rental = updateRentalRent(t, c, rental, 2500) // → 1 history row (updated)
	if err := c.DeleteRental(ctx, rental.ID); err != nil {
		t.Fatalf("DeleteRental: %v", err)
	} // → 1 more history row (deleted)

	history, err := c.GetRentalHistoryByRentalID(ctx, rental.ID)
	if err != nil {
		t.Fatalf("GetRentalHistoryByRentalID: %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("len = %d, want 2 (1 update + 1 delete)", len(history))
	}

	reasons := map[string]bool{}
	for _, row := range history {
		reasons[row.ArchiveReason] = true
	}
	if !reasons["updated"] {
		t.Error("missing updated row in history")
	}
	if !reasons["deleted"] {
		t.Error("missing deleted row in history")
	}
}

func TestRentalHistory_ManualArchivePlusTrigger_NoDuplicate(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	rental := seedRental(t, c, prop.ID)

	// Manually archive (e.g. with reason "rented")
	if _, err := c.ArchiveRental(ctx, client.ArchiveRentalParams{
		RentalID:         rental.ID,
		PropertyID:       prop.ID,
		SourceName:       rental.SourceName,
		SourceUrl:        rental.SourceUrl,
		ListingRef:       rental.ListingRef,
		UnitNumber:       rental.UnitNumber,
		MonthlyRent:      rental.MonthlyRent,
		SecurityDeposit:  rental.SecurityDeposit,
		Bedrooms:         rental.Bedrooms,
		Bathrooms:        rental.Bathrooms,
		Sqft:             rental.Sqft,
		AvailableDate:    rental.AvailableDate,
		LeaseTerm:        rental.LeaseTerm,
		Status:           "rented",
		DaysOnMarket:     rental.DaysOnMarket,
		ListingCreatedAt: rental.CreatedAt,
		ListingExpiresAt: rental.ExpiresAt,
		ArchiveReason:    "rented",
	}); err != nil {
		t.Fatalf("ArchiveRental: %v", err)
	}

	// Now delete — BEFORE DELETE trigger also tries to INSERT OR IGNORE
	if err := c.DeleteRental(ctx, rental.ID); err != nil {
		t.Fatalf("DeleteRental: %v", err)
	}

	history, err := c.GetRentalHistoryByRentalID(ctx, rental.ID)
	if err != nil {
		t.Fatalf("GetRentalHistoryByRentalID: %v", err)
	}
	// Trigger uses INSERT OR IGNORE — so no duplicate if manual archive already ran
	// We expect 2: the manual "rented" + the trigger "deleted"
	if len(history) != 2 {
		t.Errorf("len = %d, want 2 (manual + trigger delete)", len(history))
	}
}

// -- Isolation ----------------------------------------------------------------

func TestRentalHistory_IsolatedPerRental(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	r1 := seedRental(t, c, prop.ID, func(p *client.CreateRentalParams) { p.UnitNumber = "1A" })
	r2 := seedRental(t, c, prop.ID, func(p *client.CreateRentalParams) { p.UnitNumber = "2B" })

	updateRentalRent(t, c, r1, 2100)
	updateRentalRent(t, c, r1, 2200)
	updateRentalRent(t, c, r2, 3000)

	h1, err := c.GetRentalHistoryByRentalID(ctx, r1.ID)
	if err != nil {
		t.Fatalf("r1 history: %v", err)
	}
	h2, err := c.GetRentalHistoryByRentalID(ctx, r2.ID)
	if err != nil {
		t.Fatalf("r2 history: %v", err)
	}

	if len(h1) != 2 {
		t.Errorf("r1 history len = %d, want 2", len(h1))
	}
	if len(h2) != 1 {
		t.Errorf("r2 history len = %d, want 1", len(h2))
	}
}
