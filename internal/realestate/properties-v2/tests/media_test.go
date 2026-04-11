package tests

import (
	"testing"

	"sqlite-realestate/client"
)

func seedPhoto(t *testing.T, c *client.Client, propertyID string, overrides ...func(*client.AddPhotoParams)) client.PropertyPhoto {
	t.Helper()
	p := client.AddPhotoParams{
		PropertyID: propertyID,
		ListingID:  nil,
		RentalID:   nil,
		SourceUrl:  "https://photos.example.com/img.jpg",
		LocalPath:  "",
		Caption:    "Test photo",
		IsPrimary:  0,
		Width:      1920,
		Height:     1080,
		SizeBytes:  500000,
		MimeType:   "image/jpeg",
		SortOrder:  0,
	}
	for _, fn := range overrides {
		fn(&p)
	}
	photo, err := c.AddPhoto(ctx, p)
	if err != nil {
		t.Fatalf("AddPhoto: %v", err)
	}
	return photo
}

func TestAddPhoto(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	photo := seedPhoto(t, c, prop.ID)

	if photo.ID == "" {
		t.Error("expected non-empty ID")
	}
	if photo.PropertyID != prop.ID {
		t.Errorf("PropertyID = %q, want %q", photo.PropertyID, prop.ID)
	}
}

func TestGetPhoto(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	created := seedPhoto(t, c, prop.ID)

	got, err := c.GetPhoto(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetPhoto: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID = %q, want %q", got.ID, created.ID)
	}
}

func TestListPhotosForProperty(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)

	for i := range 3 {
		url := []string{"https://ex.com/a.jpg", "https://ex.com/b.jpg", "https://ex.com/c.jpg"}[i]
		seedPhoto(t, c, prop.ID, func(p *client.AddPhotoParams) { p.SourceUrl = url })
	}

	photos, err := c.ListPhotosForProperty(ctx, prop.ID)
	if err != nil {
		t.Fatalf("ListPhotosForProperty: %v", err)
	}
	if len(photos) != 3 {
		t.Errorf("len = %d, want 3", len(photos))
	}
}

func TestSetPrimaryPhoto(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)

	p1 := seedPhoto(t, c, prop.ID, func(p *client.AddPhotoParams) {
		p.SourceUrl = "https://ex.com/p1.jpg"
		p.IsPrimary = 1
	})
	p2 := seedPhoto(t, c, prop.ID, func(p *client.AddPhotoParams) {
		p.SourceUrl = "https://ex.com/p2.jpg"
	})

	// Clear primary on all, then mark p2
	if err := c.SetPrimaryPhoto(ctx, prop.ID); err != nil {
		t.Fatalf("SetPrimaryPhoto: %v", err)
	}
	updated, err := c.MarkPhotoAsPrimary(ctx, p2.ID)
	if err != nil {
		t.Fatalf("MarkPhotoAsPrimary: %v", err)
	}
	if updated.IsPrimary != 1 {
		t.Errorf("IsPrimary = %d, want 1", updated.IsPrimary)
	}

	// p1 should no longer be primary
	got, err := c.GetPhoto(ctx, p1.ID)
	if err != nil {
		t.Fatalf("GetPhoto p1: %v", err)
	}
	if got.IsPrimary != 0 {
		t.Errorf("p1 IsPrimary = %d, want 0 after SetPrimaryPhoto", got.IsPrimary)
	}
}

func TestGetPrimaryPhotoForProperty(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	primary := seedPhoto(t, c, prop.ID, func(p *client.AddPhotoParams) {
		p.IsPrimary = 1
		p.SourceUrl = "https://ex.com/primary.jpg"
	})
	seedPhoto(t, c, prop.ID, func(p *client.AddPhotoParams) {
		p.SourceUrl = "https://ex.com/other.jpg"
	})

	got, err := c.GetPrimaryPhotoForProperty(ctx, prop.ID)
	if err != nil {
		t.Fatalf("GetPrimaryPhotoForProperty: %v", err)
	}
	if got.ID != primary.ID {
		t.Errorf("ID = %q, want %q", got.ID, primary.ID)
	}
}

func TestDeletePhoto(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	photo := seedPhoto(t, c, prop.ID)

	if err := c.DeletePhoto(ctx, photo.ID); err != nil {
		t.Fatalf("DeletePhoto: %v", err)
	}
	_, err := c.GetPhoto(ctx, photo.ID)
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}

func TestDeletePhotosForProperty(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	for _, url := range []string{"https://ex.com/x.jpg", "https://ex.com/y.jpg"} {
		seedPhoto(t, c, prop.ID, func(p *client.AddPhotoParams) { p.SourceUrl = url })
	}

	if err := c.DeletePhotosForProperty(ctx, prop.ID); err != nil {
		t.Fatalf("DeletePhotosForProperty: %v", err)
	}
	photos, err := c.ListPhotosForProperty(ctx, prop.ID)
	if err != nil {
		t.Fatalf("ListPhotosForProperty: %v", err)
	}
	if len(photos) != 0 {
		t.Errorf("len = %d, want 0 after delete", len(photos))
	}
}

func TestUpsertListingSource(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)

	src, err := c.UpsertListingSource(ctx, client.UpsertListingSourceParams{
		PropertyID:  prop.ID,
		SourceName:  "Zillow",
		SourceUrl:   "https://zillow.com/test",
		SourceType:  "sale",
		LastSeenAt:  "2026-04-01T00:00:00Z",
		FirstSeenAt: "2026-03-01T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("UpsertListingSource: %v", err)
	}
	if src.ID == "" {
		t.Error("expected non-empty ID")
	}
	if src.SourceName != "Zillow" {
		t.Errorf("SourceName = %q, want Zillow", src.SourceName)
	}
}

func TestUpsertListingSource_UpdatesLastSeen(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)

	params := client.UpsertListingSourceParams{
		PropertyID:  prop.ID,
		SourceName:  "Redfin",
		SourceUrl:   "https://redfin.com/test",
		SourceType:  "sale",
		LastSeenAt:  "2026-03-01T00:00:00Z",
		FirstSeenAt: "2026-03-01T00:00:00Z",
	}
	if _, err := c.UpsertListingSource(ctx, params); err != nil {
		t.Fatalf("first upsert: %v", err)
	}

	// Second upsert — same URL, updated last_seen_at
	params.LastSeenAt = "2026-04-04T00:00:00Z"
	updated, err := c.UpsertListingSource(ctx, params)
	if err != nil {
		t.Fatalf("second upsert: %v", err)
	}
	if updated.LastSeenAt != "2026-04-04T00:00:00Z" {
		t.Errorf("LastSeenAt = %q, want updated value", updated.LastSeenAt)
	}
	// first_seen_at should not have changed
	if updated.FirstSeenAt != "2026-03-01T00:00:00Z" {
		t.Errorf("FirstSeenAt changed to %q, should be stable", updated.FirstSeenAt)
	}
}

func TestGetListingSources(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)

	for _, src := range []struct{ name, url string }{
		{"Zillow", "https://zillow.com/test"},
		{"Redfin", "https://redfin.com/test"},
	} {
		if _, err := c.UpsertListingSource(ctx, client.UpsertListingSourceParams{
			PropertyID:  prop.ID,
			SourceName:  src.name,
			SourceUrl:   src.url,
			SourceType:  "sale",
			LastSeenAt:  "2026-04-01T00:00:00Z",
			FirstSeenAt: "2026-04-01T00:00:00Z",
		}); err != nil {
			t.Fatalf("UpsertListingSource %s: %v", src.name, err)
		}
	}

	sources, err := c.GetListingSources(ctx, prop.ID)
	if err != nil {
		t.Fatalf("GetListingSources: %v", err)
	}
	if len(sources) != 2 {
		t.Errorf("len = %d, want 2", len(sources))
	}
}

func TestDeactivateListingSource(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	src, err := c.UpsertListingSource(ctx, client.UpsertListingSourceParams{
		PropertyID:  prop.ID,
		SourceName:  "Zillow",
		SourceUrl:   "https://zillow.com/test",
		SourceType:  "sale",
		LastSeenAt:  "2026-04-01T00:00:00Z",
		FirstSeenAt: "2026-04-01T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("UpsertListingSource: %v", err)
	}

	if err := c.DeactivateListingSource(ctx, src.ID); err != nil {
		t.Fatalf("DeactivateListingSource: %v", err)
	}

	sources, err := c.GetListingSources(ctx, prop.ID)
	if err != nil {
		t.Fatalf("GetListingSources: %v", err)
	}
	if len(sources) != 1 || sources[0].IsActive != 0 {
		t.Errorf("expected 1 deactivated source, got %d active=%v", len(sources), sources[0].IsActive)
	}
}
