package tests

import (
	"testing"

	"sqlite-realestate/client"
)

func seedProperty(t *testing.T, c *client.Client, overrides ...func(*client.CreatePropertyParams)) client.Property {
	t.Helper()
	p := client.CreatePropertyParams{
		Address:       "123 Test St",
		City:          "Los Angeles",
		State:         "CA",
		ZipCode:       "90001",
		County:        "Los Angeles",
		Lat:           34.052,
		Lng:           -118.243,
		PropertyName:  "Test House",
		PropertyType:  "sfr",
		Bedrooms:      3,
		Bathrooms:     2,
		Sqft:          1500,
		BuildingSf:    1500,
		LotSf:         6000,
		YearBuilt:     1990,
		NumberOfUnits: 1,
		Organization:  "Test Org",
		Notes:         "seed property",
	}
	for _, fn := range overrides {
		fn(&p)
	}
	prop, err := c.CreateProperty(ctx, p)
	if err != nil {
		t.Fatalf("CreateProperty: %v", err)
	}
	return client.WrapProperty(prop)
}

func TestCreateProperty(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedProperty(t, c)

	if prop.ID == "" {
		t.Error("expected non-empty ID")
	}
	if prop.Address != "123 Test St" {
		t.Errorf("address = %q, want %q", prop.Address, "123 Test St")
	}
	if prop.CreatedAt == "" {
		t.Error("expected CreatedAt to be set")
	}
	if prop.UpdatedAt == "" {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestGetProperty(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	created := seedProperty(t, c)

	got, err := c.GetProperty(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetProperty: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID = %q, want %q", got.ID, created.ID)
	}
}

func TestGetPropertyByAddress(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	created := seedProperty(t, c)

	got, err := c.GetPropertyByAddress(ctx, created.Address)
	if err != nil {
		t.Fatalf("GetPropertyByAddress: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("got ID %q, want %q", got.ID, created.ID)
	}
}

func TestGetPropertyNotFound(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	_, err := c.GetProperty(ctx, "nonexistent-id")
	if err == nil {
		t.Error("expected error for missing property, got nil")
	}
}

func TestUpdateProperty(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	created := seedProperty(t, c)

	updated, err := c.UpdateProperty(ctx, client.UpdatePropertyParams{
		ID:            created.ID,
		Address:       created.Address,
		City:          "San Francisco",
		State:         "CA",
		ZipCode:       "94105",
		County:        "San Francisco",
		Lat:           created.Lat,
		Lng:           created.Lng,
		PropertyName:  "Updated Name",
		PropertyType:  created.PropertyType,
		Bedrooms:      created.Bedrooms,
		Bathrooms:     created.Bathrooms,
		Sqft:          created.Sqft,
		BuildingSf:    created.BuildingSf,
		LotSf:         created.LotSf,
		YearBuilt:     created.YearBuilt,
		NumberOfUnits: created.NumberOfUnits,
		Organization:  created.Organization,
		Notes:         created.Notes,
	})
	if err != nil {
		t.Fatalf("UpdateProperty: %v", err)
	}
	if updated.City != "San Francisco" {
		t.Errorf("City = %q, want San Francisco", updated.City)
	}
	if updated.PropertyName != "Updated Name" {
		t.Errorf("PropertyName = %q, want Updated Name", updated.PropertyName)
	}
}

func TestUpdatePropertyCoords(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	created := seedProperty(t, c)

	updated, err := c.UpdatePropertyCoords(ctx, client.UpdatePropertyCoordsParams{
		ID:  created.ID,
		Lat: 37.790,
		Lng: -122.399,
	})
	if err != nil {
		t.Fatalf("UpdatePropertyCoords: %v", err)
	}
	if updated.Lat != 37.790 {
		t.Errorf("Lat = %v, want 37.790", updated.Lat)
	}
	if updated.Lng != -122.399 {
		t.Errorf("Lng = %v, want -122.399", updated.Lng)
	}
}

func TestDeleteProperty(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	created := seedProperty(t, c)

	if err := c.DeleteProperty(ctx, created.ID); err != nil {
		t.Fatalf("DeleteProperty: %v", err)
	}

	_, err := c.GetProperty(ctx, created.ID)
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}

func TestListProperties(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	for _, addr := range []string{"111 A St", "222 B St", "333 C St"} {
		seedProperty(t, c, func(p *client.CreatePropertyParams) { p.Address = addr })
	}

	list, err := c.ListProperties(ctx, client.ListPropertiesParams{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("ListProperties: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("len = %d, want 3", len(list))
	}
}

func TestListPropertiesPagination(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	for _, addr := range []string{"1 A", "2 B", "3 C", "4 D", "5 E"} {
		seedProperty(t, c, func(p *client.CreatePropertyParams) { p.Address = addr })
	}

	page1, err := c.ListProperties(ctx, client.ListPropertiesParams{Limit: 3, Offset: 0})
	if err != nil {
		t.Fatalf("page1: %v", err)
	}
	page2, err := c.ListProperties(ctx, client.ListPropertiesParams{Limit: 3, Offset: 3})
	if err != nil {
		t.Fatalf("page2: %v", err)
	}
	if len(page1) != 3 {
		t.Errorf("page1 len = %d, want 3", len(page1))
	}
	if len(page2) != 2 {
		t.Errorf("page2 len = %d, want 2", len(page2))
	}
}

func TestListPropertiesByCityState(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	seedProperty(t, c, func(p *client.CreatePropertyParams) {
		p.Address = "100 LA St"
		p.City = "Los Angeles"
		p.State = "CA"
	})
	seedProperty(t, c, func(p *client.CreatePropertyParams) {
		p.Address = "200 SF St"
		p.City = "San Francisco"
		p.State = "CA"
	})

	results, err := c.ListPropertiesByCityState(ctx, client.ListPropertiesByCityStateParams{
		City: "Los Angeles", State: "CA", Limit: 10, Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListPropertiesByCityState: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("len = %d, want 1", len(results))
	}
	if results[0].City != "Los Angeles" {
		t.Errorf("City = %q, want Los Angeles", results[0].City)
	}
}

func TestCountProperties(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	for _, addr := range []string{"1 X", "2 Y", "3 Z"} {
		seedProperty(t, c, func(p *client.CreatePropertyParams) { p.Address = addr })
	}

	count, err := c.CountProperties(ctx)
	if err != nil {
		t.Fatalf("CountProperties: %v", err)
	}
	if count != 3 {
		t.Errorf("count = %d, want 3", count)
	}
}

func TestAddressUniqueConstraint(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	seedProperty(t, c)

	_, err := c.CreateProperty(ctx, client.CreatePropertyParams{
		Address: "123 Test St", // duplicate
		City:    "Los Angeles",
		State:   "CA",
	})
	if err == nil {
		t.Error("expected unique constraint error for duplicate address, got nil")
	}
}
