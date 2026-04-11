package tests

import (
	"testing"

	"sqlite-realestate/client"
)

func seedRental(t *testing.T, c *client.Client, propertyID string, overrides ...func(*client.CreateRentalParams)) client.RentalListing {
	t.Helper()
	p := client.CreateRentalParams{
		PropertyID:      propertyID,
		SourceName:      "Apartments.com",
		SourceUrl:       "https://apartments.com/test",
		ListingRef:      "APT-001",
		MonthlyRent:     2000,
		SecurityDeposit: 2000,
		RentPerSqft:     2.0,
		UnitNumber:      "",
		Bedrooms:        2,
		Bathrooms:       1,
		Sqft:            1000,
		AvailableDate:   "2026-05-01",
		LeaseTerm:       "12mo",
		PetsAllowed:     0,
		Furnished:       0,
		Status:          "active",
		DaysOnMarket:    3,
		ExpiresAt:       futureDate(30),
	}
	for _, fn := range overrides {
		fn(&p)
	}
	rental, err := c.CreateRental(ctx, p)
	if err != nil {
		t.Fatalf("CreateRental: %v", err)
	}
	return client.WrapRental(rental)
}

func TestCreateRental(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	rental := seedRental(t, c, prop.ID)

	if rental.ID == "" {
		t.Error("expected non-empty ID")
	}
	if rental.PropertyID != prop.ID {
		t.Errorf("PropertyID = %q, want %q", rental.PropertyID, prop.ID)
	}
	if rental.MonthlyRent != 2000 {
		t.Errorf("MonthlyRent = %v, want 2000", rental.MonthlyRent)
	}
}

func TestGetRental(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	created := seedRental(t, c, prop.ID)

	got, err := c.GetRental(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetRental: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID = %q, want %q", got.ID, created.ID)
	}
}

func TestGetActiveRentalByPropertyUnit(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	created := seedRental(t, c, prop.ID, func(p *client.CreateRentalParams) {
		p.UnitNumber = "2B"
	})

	got, err := c.GetActiveRentalByPropertyUnit(ctx, client.GetActiveRentalByPropertyUnitParams{
		PropertyID: prop.ID,
		UnitNumber: "2B",
	})
	if err != nil {
		t.Fatalf("GetActiveRentalByPropertyUnit: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID = %q, want %q", got.ID, created.ID)
	}
}

func TestOneRentalPerPropertyUnit(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	seedRental(t, c, prop.ID, func(p *client.CreateRentalParams) { p.UnitNumber = "1A" })

	_, err := c.CreateRental(ctx, client.CreateRentalParams{
		PropertyID: prop.ID,
		UnitNumber: "1A", // duplicate (property_id, unit_number)
		Status:     "active",
		ExpiresAt:  futureDate(30),
	})
	if err == nil {
		t.Error("expected unique constraint error for duplicate (property, unit), got nil")
	}
}

func TestUpdateRentalTTL(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	rental := seedRental(t, c, prop.ID)

	newExpiry := futureDate(60)
	updated, err := c.UpdateRentalTTL(ctx, client.UpdateRentalTTLParams{
		ExpiresAt: newExpiry,
		ID:        rental.ID,
	})
	if err != nil {
		t.Fatalf("UpdateRentalTTL: %v", err)
	}
	if updated.ExpiresAt != newExpiry {
		t.Errorf("ExpiresAt = %q, want %q", updated.ExpiresAt, newExpiry)
	}
}

func TestUpdateRentalStatus(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	rental := seedRental(t, c, prop.ID)

	updated, err := c.UpdateRentalStatus(ctx, client.UpdateRentalStatusParams{
		Status: "rented",
		ID:     rental.ID,
	})
	if err != nil {
		t.Fatalf("UpdateRentalStatus: %v", err)
	}
	if updated.Status != "rented" {
		t.Errorf("Status = %q, want rented", updated.Status)
	}
}

func TestListActiveRentals(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	for i, addr := range []string{"1 Rent St", "2 Rent Ave", "3 Rent Rd"} {
		prop := seedProperty(t, c, func(p *client.CreatePropertyParams) { p.Address = addr })
		unit := []string{"A", "B", "C"}[i]
		seedRental(t, c, prop.ID, func(p *client.CreateRentalParams) { p.UnitNumber = unit })
	}

	list, err := c.ListActiveRentals(ctx, client.ListActiveRentalsParams{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("ListActiveRentals: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("len = %d, want 3", len(list))
	}
}

func TestListActiveRentalsByPriceRange(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	low := seedProperty(t, c, func(p *client.CreatePropertyParams) { p.Address = "1 Low" })
	mid := seedProperty(t, c, func(p *client.CreatePropertyParams) { p.Address = "2 Mid" })
	high := seedProperty(t, c, func(p *client.CreatePropertyParams) { p.Address = "3 High" })

	seedRental(t, c, low.ID, func(p *client.CreateRentalParams) { p.MonthlyRent = 1500 })
	seedRental(t, c, mid.ID, func(p *client.CreateRentalParams) { p.MonthlyRent = 2500 })
	seedRental(t, c, high.ID, func(p *client.CreateRentalParams) { p.MonthlyRent = 4000 })

	results, err := c.ListActiveRentalsByPriceRange(ctx, client.ListActiveRentalsByPriceRangeParams{
		MonthlyRent:   1000,
		MonthlyRent_2: 3000,
		Limit:         10,
		Offset:        0,
	})
	if err != nil {
		t.Fatalf("ListActiveRentalsByPriceRange: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("len = %d, want 2 (1500 and 2500)", len(results))
	}
}

func TestListExpiredRentals(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	propA := seedProperty(t, c, func(p *client.CreatePropertyParams) { p.Address = "1 Active" })
	propB := seedProperty(t, c, func(p *client.CreatePropertyParams) { p.Address = "2 Expired" })

	seedRental(t, c, propA.ID)
	seedRental(t, c, propB.ID, func(p *client.CreateRentalParams) {
		p.ExpiresAt = pastDate(1)
	})

	expired, err := c.ListExpiredRentals(ctx)
	if err != nil {
		t.Fatalf("ListExpiredRentals: %v", err)
	}
	if len(expired) != 1 {
		t.Errorf("len = %d, want 1", len(expired))
	}
}

func TestArchiveRental(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	rental := seedRental(t, c, prop.ID)

	archived, err := c.ArchiveRental(ctx, client.ArchiveRentalParams{
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
	})
	if err != nil {
		t.Fatalf("ArchiveRental: %v", err)
	}
	if archived.RentalID != rental.ID {
		t.Errorf("RentalID = %q, want %q", archived.RentalID, rental.ID)
	}

	if err := c.DeleteRental(ctx, rental.ID); err != nil {
		t.Fatalf("DeleteRental: %v", err)
	}

	history, err := c.GetRentalHistoryByRentalID(ctx, rental.ID)
	if err != nil {
		t.Fatalf("GetRentalHistoryByRentalID: %v", err)
	}
	// 2 rows: manual "rented" archive + auto "deleted" from BEFORE DELETE trigger
	if len(history) != 2 {
		t.Errorf("history len = %d, want 2", len(history))
	}
	reasons := map[string]bool{}
	for _, h := range history {
		reasons[h.ArchiveReason] = true
	}
	if !reasons["rented"] {
		t.Error("missing manual rented archive row")
	}
	if !reasons["deleted"] {
		t.Error("missing auto-deleted trigger row")
	}
}

func TestDeleteRental(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	rental := seedRental(t, c, prop.ID)

	if err := c.DeleteRental(ctx, rental.ID); err != nil {
		t.Fatalf("DeleteRental: %v", err)
	}
	_, err := c.GetRental(ctx, rental.ID)
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}
