package tests

import (
	"testing"

	"sqlite-realestate/client"
)

// Verifies that the properties_history table receives a snapshot
// whenever a property row is created, updated, or deleted.
//
// Trigger source: 006_history_triggers.sql
//   - trg_properties_history_update  → fires AFTER UPDATE
//   - trg_properties_history_delete  → fires AFTER DELETE
//
// Note: INSERT does not write history (no snapshot exists yet).

// helper: update a property's city and return the new row
func updatePropertyCity(t *testing.T, c *client.Client, prop client.Property, city string) client.Property {
	t.Helper()
	updated, err := c.UpdateProperty(ctx, client.UpdatePropertyParams{
		ID:            prop.ID,
		Address:       prop.Address,
		City:          city,
		State:         prop.State,
		ZipCode:       prop.ZipCode,
		County:        prop.County,
		Lat:           prop.Lat,
		Lng:           prop.Lng,
		PropertyName:  prop.PropertyName,
		PropertyType:  prop.PropertyType,
		Bedrooms:      prop.Bedrooms,
		Bathrooms:     prop.Bathrooms,
		Sqft:          prop.Sqft,
		BuildingSf:    prop.BuildingSf,
		LotSf:         prop.LotSf,
		YearBuilt:     prop.YearBuilt,
		NumberOfUnits: prop.NumberOfUnits,
		Organization:  prop.Organization,
		Notes:         prop.Notes,
	})
	if err != nil {
		t.Fatalf("UpdateProperty: %v", err)
	}
	return client.WrapProperty(updated)
}

// -- INSERT -------------------------------------------------------------------

func TestPropertiesHistory_NoRowOnInsert(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)

	history, err := c.GetPropertyHistory(ctx, client.GetPropertyHistoryParams{
		PropertyID: prop.ID,
		Limit:      10,
		Offset:     0,
	})
	if err != nil {
		t.Fatalf("GetPropertyHistory: %v", err)
	}
	if len(history) != 0 {
		t.Errorf("len = %d, want 0 — insert should not write history", len(history))
	}
}

// -- UPDATE -------------------------------------------------------------------

func TestPropertiesHistory_SingleUpdate(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	updatePropertyCity(t, c, prop, "San Francisco")

	history, err := c.GetPropertyHistory(ctx, client.GetPropertyHistoryParams{
		PropertyID: prop.ID,
		Limit:      10,
		Offset:     0,
	})
	if err != nil {
		t.Fatalf("GetPropertyHistory: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("len = %d, want 1", len(history))
	}

	row := history[0]
	if row.ChangeType != "updated" {
		t.Errorf("ChangeType = %q, want updated", row.ChangeType)
	}
	// Snapshot must hold the OLD city value
	if row.City != "Los Angeles" {
		t.Errorf("snapshot City = %q, want Los Angeles (pre-update value)", row.City)
	}
	if row.PropertyID != prop.ID {
		t.Errorf("PropertyID = %q, want %q", row.PropertyID, prop.ID)
	}
	if row.ChangedAt == "" {
		t.Error("ChangedAt must be set")
	}
}

func TestPropertiesHistory_MultipleUpdates_OneRowEach(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)

	cities := []string{"Denver", "Chicago", "Miami"}
	for _, city := range cities {
		prop = updatePropertyCity(t, c, prop, city)
	}

	history, err := c.GetPropertyHistory(ctx, client.GetPropertyHistoryParams{
		PropertyID: prop.ID,
		Limit:      10,
		Offset:     0,
	})
	if err != nil {
		t.Fatalf("GetPropertyHistory: %v", err)
	}
	if len(history) != 3 {
		t.Errorf("len = %d, want 3 (one snapshot per update)", len(history))
	}
	for _, row := range history {
		if row.ChangeType != "updated" {
			t.Errorf("ChangeType = %q, want updated", row.ChangeType)
		}
	}
}

func TestPropertiesHistory_SnapshotPreservesAllFields(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	updatePropertyCity(t, c, prop, "Portland")

	history, err := c.GetPropertyHistory(ctx, client.GetPropertyHistoryParams{
		PropertyID: prop.ID,
		Limit:      1,
		Offset:     0,
	})
	if err != nil {
		t.Fatalf("GetPropertyHistory: %v", err)
	}
	if len(history) == 0 {
		t.Fatal("no history rows")
	}

	row := history[0]
	if row.Address != prop.Address {
		t.Errorf("Address = %q, want %q", row.Address, prop.Address)
	}
	if row.State != prop.State {
		t.Errorf("State = %q, want %q", row.State, prop.State)
	}
	if row.Lat != prop.Lat {
		t.Errorf("Lat = %v, want %v", row.Lat, prop.Lat)
	}
	if row.Lng != prop.Lng {
		t.Errorf("Lng = %v, want %v", row.Lng, prop.Lng)
	}
	if row.Bedrooms != prop.Bedrooms {
		t.Errorf("Bedrooms = %v, want %v", row.Bedrooms, prop.Bedrooms)
	}
	if row.Organization != prop.Organization {
		t.Errorf("Organization = %q, want %q", row.Organization, prop.Organization)
	}
}

func TestPropertiesHistory_CoordsUpdate(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)

	if _, err := c.UpdatePropertyCoords(ctx, client.UpdatePropertyCoordsParams{
		ID:  prop.ID,
		Lat: 37.790,
		Lng: -122.399,
	}); err != nil {
		t.Fatalf("UpdatePropertyCoords: %v", err)
	}

	history, err := c.GetPropertyHistory(ctx, client.GetPropertyHistoryParams{
		PropertyID: prop.ID,
		Limit:      10,
		Offset:     0,
	})
	if err != nil {
		t.Fatalf("GetPropertyHistory: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("len = %d, want 1", len(history))
	}
	// Snapshot holds OLD coords
	if history[0].Lat != prop.Lat {
		t.Errorf("snapshot Lat = %v, want %v (old)", history[0].Lat, prop.Lat)
	}
}

// -- DELETE -------------------------------------------------------------------

func TestPropertiesHistory_Delete(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)

	if err := c.DeleteProperty(ctx, prop.ID); err != nil {
		t.Fatalf("DeleteProperty: %v", err)
	}

	history, err := c.GetPropertyHistory(ctx, client.GetPropertyHistoryParams{
		PropertyID: prop.ID,
		Limit:      10,
		Offset:     0,
	})
	if err != nil {
		t.Fatalf("GetPropertyHistory: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("len = %d, want 1", len(history))
	}

	row := history[0]
	if row.ChangeType != "deleted" {
		t.Errorf("ChangeType = %q, want deleted", row.ChangeType)
	}
	if row.Address != prop.Address {
		t.Errorf("snapshot Address = %q, want %q", row.Address, prop.Address)
	}
}

func TestPropertiesHistory_UpdateThenDelete(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	updatePropertyCity(t, c, prop, "Austin")

	if err := c.DeleteProperty(ctx, prop.ID); err != nil {
		t.Fatalf("DeleteProperty: %v", err)
	}

	history, err := c.GetPropertyHistory(ctx, client.GetPropertyHistoryParams{
		PropertyID: prop.ID,
		Limit:      10,
		Offset:     0,
	})
	if err != nil {
		t.Fatalf("GetPropertyHistory: %v", err)
	}
	// 1 update + 1 delete = 2 rows
	if len(history) != 2 {
		t.Fatalf("len = %d, want 2", len(history))
	}
}

// -- FilterByChangeType -------------------------------------------------------

func TestPropertiesHistory_FilterUpdated(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	updatePropertyCity(t, c, prop, "Seattle")
	if err := c.DeleteProperty(ctx, prop.ID); err != nil {
		t.Fatalf("DeleteProperty: %v", err)
	}

	rows, err := c.GetPropertyHistoryByChangeType(ctx, client.GetPropertyHistoryByChangeTypeParams{
		PropertyID: prop.ID,
		ChangeType: "updated",
	})
	if err != nil {
		t.Fatalf("GetPropertyHistoryByChangeType: %v", err)
	}
	if len(rows) != 1 {
		t.Errorf("len = %d, want 1", len(rows))
	}
	if rows[0].ChangeType != "updated" {
		t.Errorf("ChangeType = %q, want updated", rows[0].ChangeType)
	}
}

func TestPropertiesHistory_FilterDeleted(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)
	updatePropertyCity(t, c, prop, "Seattle")
	if err := c.DeleteProperty(ctx, prop.ID); err != nil {
		t.Fatalf("DeleteProperty: %v", err)
	}

	rows, err := c.GetPropertyHistoryByChangeType(ctx, client.GetPropertyHistoryByChangeTypeParams{
		PropertyID: prop.ID,
		ChangeType: "deleted",
	})
	if err != nil {
		t.Fatalf("GetPropertyHistoryByChangeType: %v", err)
	}
	if len(rows) != 1 {
		t.Errorf("len = %d, want 1", len(rows))
	}
	if rows[0].ChangeType != "deleted" {
		t.Errorf("ChangeType = %q, want deleted", rows[0].ChangeType)
	}
}

// -- Isolation: other properties' history is not mixed in ---------------------

func TestPropertiesHistory_IsolatedPerProperty(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	p1 := seedProperty(t, c, func(p *client.CreatePropertyParams) { p.Address = "1 Iso St" })
	p2 := seedProperty(t, c, func(p *client.CreatePropertyParams) { p.Address = "2 Iso St" })

	updatePropertyCity(t, c, p1, "Portland")
	updatePropertyCity(t, c, p1, "Boise")
	updatePropertyCity(t, c, p2, "Tucson")

	h1, err := c.GetPropertyHistory(ctx, client.GetPropertyHistoryParams{PropertyID: p1.ID, Limit: 10})
	if err != nil {
		t.Fatalf("p1 history: %v", err)
	}
	h2, err := c.GetPropertyHistory(ctx, client.GetPropertyHistoryParams{PropertyID: p2.ID, Limit: 10})
	if err != nil {
		t.Fatalf("p2 history: %v", err)
	}

	if len(h1) != 2 {
		t.Errorf("p1 history len = %d, want 2", len(h1))
	}
	if len(h2) != 1 {
		t.Errorf("p2 history len = %d, want 1", len(h2))
	}
}
