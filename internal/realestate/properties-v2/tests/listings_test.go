package tests

import (
	"testing"

	"sqlite-realestate/client"
)

func seedListing(t *testing.T, c *client.Client, propertyID string, overrides ...func(*client.CreateListingParams)) client.PropertyListing {
	t.Helper()
	p := client.CreateListingParams{
		PropertyID:   propertyID,
		SourceName:   "Zillow",
		SourceUrl:    "https://zillow.com/test",
		MlsID:        "MLS-TEST-001",
		ListPrice:    500000,
		PricePerSqft: 333.33,
		Status:       "active",
		DaysOnMarket: 5,
		ExpiresAt:    futureDate(30),
	}
	for _, fn := range overrides {
		fn(&p)
	}
	listing, err := c.CreateListing(ctx, p)
	if err != nil {
		t.Fatalf("CreateListing: %v", err)
	}
	return client.WrapListing(listing)
}

func TestCreateListing(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	listing := seedListing(t, c, prop.ID)

	if listing.ID == "" {
		t.Error("expected non-empty ID")
	}
	if listing.PropertyID != prop.ID {
		t.Errorf("PropertyID = %q, want %q", listing.PropertyID, prop.ID)
	}
	if listing.ListPrice != 500000 {
		t.Errorf("ListPrice = %v, want 500000", listing.ListPrice)
	}
}

func TestGetListing(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	created := seedListing(t, c, prop.ID)

	got, err := c.GetListing(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetListing: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID = %q, want %q", got.ID, created.ID)
	}
}

func TestGetActiveListingByProperty(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	created := seedListing(t, c, prop.ID)

	got, err := c.GetActiveListingByProperty(ctx, prop.ID)
	if err != nil {
		t.Fatalf("GetActiveListingByProperty: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID = %q, want %q", got.ID, created.ID)
	}
}

func TestOneActiveListing_UniquePerProperty(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	seedListing(t, c, prop.ID)

	_, err := c.CreateListing(ctx, client.CreateListingParams{
		PropertyID: prop.ID, // duplicate property_id
		SourceName: "Redfin",
		Status:     "active",
		ExpiresAt:  futureDate(30),
	})
	if err == nil {
		t.Error("expected unique constraint error for duplicate property listing, got nil")
	}
}

func TestUpdateListingTTL(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	listing := seedListing(t, c, prop.ID)

	newExpiry := futureDate(60)
	updated, err := c.UpdateListingTTL(ctx, client.UpdateListingTTLParams{
		ExpiresAt: newExpiry,
		ID:        listing.ID,
	})
	if err != nil {
		t.Fatalf("UpdateListingTTL: %v", err)
	}
	if updated.ExpiresAt != newExpiry {
		t.Errorf("ExpiresAt = %q, want %q", updated.ExpiresAt, newExpiry)
	}
}

func TestUpdateListingStatus(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	listing := seedListing(t, c, prop.ID)

	updated, err := c.UpdateListingStatus(ctx, client.UpdateListingStatusParams{
		Status: "pending",
		ID:     listing.ID,
	})
	if err != nil {
		t.Fatalf("UpdateListingStatus: %v", err)
	}
	if updated.Status != "pending" {
		t.Errorf("Status = %q, want pending", updated.Status)
	}
}

func TestListActiveListings(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	addrs := []string{"10 Active St", "20 Active Ave", "30 Active Rd"}
	for _, addr := range addrs {
		prop := seedProperty(t, c, func(p *client.CreatePropertyParams) { p.Address = addr })
		seedListing(t, c, prop.ID)
	}

	list, err := c.ListActiveListings(ctx, client.ListActiveListingsParams{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("ListActiveListings: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("len = %d, want 3", len(list))
	}
}

func TestListExpiredListings(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop1 := seedProperty(t, c, func(p *client.CreatePropertyParams) { p.Address = "1 Active" })
	prop2 := seedProperty(t, c, func(p *client.CreatePropertyParams) { p.Address = "2 Expired" })

	// Active listing
	seedListing(t, c, prop1.ID)
	// Expired listing
	seedListing(t, c, prop2.ID, func(p *client.CreateListingParams) {
		p.ExpiresAt = pastDate(1)
	})

	expired, err := c.ListExpiredListings(ctx)
	if err != nil {
		t.Fatalf("ListExpiredListings: %v", err)
	}
	if len(expired) != 1 {
		t.Errorf("len = %d, want 1", len(expired))
	}
	if expired[0].PropertyID != prop2.ID {
		t.Errorf("PropertyID = %q, want %q", expired[0].PropertyID, prop2.ID)
	}
}

func TestArchiveListing(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	listing := seedListing(t, c, prop.ID)

	archived, err := c.ArchiveListing(ctx, client.ArchiveListingParams{
		ListingID:        listing.ID,
		PropertyID:       prop.ID,
		SourceName:       listing.SourceName,
		SourceUrl:        listing.SourceUrl,
		MlsID:            listing.MlsID,
		ListPrice:        listing.ListPrice,
		PricePerSqft:     listing.PricePerSqft,
		Status:           "sold",
		DaysOnMarket:     listing.DaysOnMarket,
		ListingCreatedAt: listing.CreatedAt,
		ListingExpiresAt: listing.ExpiresAt,
		ArchiveReason:    "sold",
	})
	if err != nil {
		t.Fatalf("ArchiveListing: %v", err)
	}
	if archived.ListingID != listing.ID {
		t.Errorf("ListingID = %q, want %q", archived.ListingID, listing.ID)
	}
	if archived.ArchiveReason != "sold" {
		t.Errorf("ArchiveReason = %q, want sold", archived.ArchiveReason)
	}

	// Delete the active listing
	if err := c.DeleteListing(ctx, listing.ID); err != nil {
		t.Fatalf("DeleteListing: %v", err)
	}

	// Confirm it's gone from active
	_, err = c.GetListing(ctx, listing.ID)
	if err == nil {
		t.Error("expected error after delete, got nil")
	}

	// Confirm it's in history: manual "sold" + auto "deleted" from BEFORE DELETE trigger
	history, err := c.GetListingHistoryByListingID(ctx, listing.ID)
	if err != nil {
		t.Fatalf("GetListingHistoryByListingID: %v", err)
	}
	if len(history) != 2 {
		t.Errorf("history len = %d, want 2 (manual sold + trigger deleted)", len(history))
	}
	reasons := map[string]bool{}
	for _, h := range history {
		reasons[h.ArchiveReason] = true
	}
	if !reasons["sold"] {
		t.Error("missing manual sold archive row")
	}
	if !reasons["deleted"] {
		t.Error("missing auto-deleted trigger row")
	}
}

func TestGetListingHistory(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	listing := seedListing(t, c, prop.ID)

	_, err := c.ArchiveListing(ctx, client.ArchiveListingParams{
		ListingID:        listing.ID,
		PropertyID:       prop.ID,
		SourceName:       listing.SourceName,
		SourceUrl:        listing.SourceUrl,
		MlsID:            listing.MlsID,
		ListPrice:        listing.ListPrice,
		PricePerSqft:     listing.PricePerSqft,
		Status:           "sold",
		DaysOnMarket:     listing.DaysOnMarket,
		ListingCreatedAt: listing.CreatedAt,
		ListingExpiresAt: listing.ExpiresAt,
		ArchiveReason:    "sold",
	})
	if err != nil {
		t.Fatalf("ArchiveListing: %v", err)
	}

	history, err := c.GetListingHistory(ctx, client.GetListingHistoryParams{
		PropertyID: prop.ID,
		Limit:      10,
		Offset:     0,
	})
	if err != nil {
		t.Fatalf("GetListingHistory: %v", err)
	}
	if len(history) != 1 {
		t.Errorf("len = %d, want 1", len(history))
	}
}

func TestDeleteListing(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	listing := seedListing(t, c, prop.ID)

	if err := c.DeleteListing(ctx, listing.ID); err != nil {
		t.Fatalf("DeleteListing: %v", err)
	}
	_, err := c.GetListing(ctx, listing.ID)
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}
