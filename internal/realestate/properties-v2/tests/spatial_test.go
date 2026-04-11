package tests

import (
	"math"
	"testing"

	"sqlite-realestate/client"
)

// Real-world coordinates used across spatial tests
// All within ~50km of downtown Los Angeles (34.052, -118.243)
var (
	laCenter    = [2]float64{34.052, -118.243}  // Downtown LA — center point
	hollywood   = [2]float64{34.098, -118.327}  // ~8km NW
	santaMonica = [2]float64{34.019, -118.491}  // ~23km W
	longBeach   = [2]float64{33.770, -118.193}  // ~32km S
	sanDiego    = [2]float64{32.714, -117.173}  // ~190km SE — outside any test radius
)

// seedPropertyAt creates a property at a specific lat/lng.
func seedPropertyAt(t *testing.T, c *client.Client, addr string, lat, lng float64) client.Property {
	t.Helper()
	return seedProperty(t, c, func(p *client.CreatePropertyParams) {
		p.Address = addr
		p.Lat = lat
		p.Lng = lng
	})
}

// ---- BBoxFromPoint ----------------------------------------------------------

func TestBBoxFromPoint_ContainsCenter(t *testing.T) {
	bbox := client.BBoxFromPoint(34.052, -118.243, 10)
	if 34.052 < bbox.MinLat || 34.052 > bbox.MaxLat {
		t.Errorf("center lat not inside bbox: %+v", bbox)
	}
	if -118.243 < bbox.MinLng || -118.243 > bbox.MaxLng {
		t.Errorf("center lng not inside bbox: %+v", bbox)
	}
}

func TestBBoxFromPoint_Size(t *testing.T) {
	// 10km radius → bbox height ≈ 20km ≈ 0.18 degrees lat
	bbox := client.BBoxFromPoint(34.052, -118.243, 10)
	latSpan := bbox.MaxLat - bbox.MinLat
	if latSpan < 0.17 || latSpan > 0.20 {
		t.Errorf("lat span = %.4f, want ~0.18 for 10km radius", latSpan)
	}
}

// ---- Haversine --------------------------------------------------------------

func TestHaversine_SamePoint(t *testing.T) {
	d := client.Haversine(34.052, -118.243, 34.052, -118.243)
	if d != 0 {
		t.Errorf("distance to self = %v, want 0", d)
	}
}

func TestHaversine_KnownDistance(t *testing.T) {
	// LA downtown (34.052, -118.243) to San Diego downtown (32.714, -117.173)
	// Haversine gives ~179km; straight-line road distance is ~192km.
	d := client.Haversine(laCenter[0], laCenter[1], sanDiego[0], sanDiego[1])
	if math.Abs(d-179) > 5 {
		t.Errorf("LA→SD = %.1fkm, want ~179km", d)
	}
}

func TestHaversine_Symmetric(t *testing.T) {
	d1 := client.Haversine(laCenter[0], laCenter[1], hollywood[0], hollywood[1])
	d2 := client.Haversine(hollywood[0], hollywood[1], laCenter[0], laCenter[1])
	if math.Abs(d1-d2) > 0.001 {
		t.Errorf("haversine not symmetric: %.4f vs %.4f", d1, d2)
	}
}

// ---- Trigger: R*Tree auto-populated on insert --------------------------------

func TestRTreePopulatedOnInsert(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedPropertyAt(t, c, "1 Trigger Test", laCenter[0], laCenter[1])

	// The insert trigger should have created a property_spatial_map row
	var rid int64
	err := c.RawDB().QueryRowContext(ctx,
		`SELECT rid FROM property_spatial_map WHERE property_id = ?`, prop.ID,
	).Scan(&rid)
	if err != nil {
		t.Fatalf("spatial_map row not found after insert: %v", err)
	}

	// And a matching rtree row
	var minLat, maxLat, minLng, maxLng float64
	err = c.RawDB().QueryRowContext(ctx,
		`SELECT min_lat, max_lat, min_lng, max_lng FROM property_rtree WHERE id = ?`, rid,
	).Scan(&minLat, &maxLat, &minLng, &maxLng)
	if err != nil {
		t.Fatalf("rtree row not found: %v", err)
	}
	// R*Tree stores floats as 32-bit internally, so compare with tolerance
	if math.Abs(minLat-laCenter[0]) > 1e-4 {
		t.Errorf("rtree min_lat = %v, want ~%v", minLat, laCenter[0])
	}
}

func TestRTreeUpdatedOnCoordsChange(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedPropertyAt(t, c, "1 Coord Update", laCenter[0], laCenter[1])

	// Move property to Hollywood coords
	if _, err := c.UpdatePropertyCoords(ctx, client.UpdatePropertyCoordsParams{
		ID: prop.ID, Lat: hollywood[0], Lng: hollywood[1],
	}); err != nil {
		t.Fatalf("UpdatePropertyCoords: %v", err)
	}

	var rid int64
	if err := c.RawDB().QueryRowContext(ctx,
		`SELECT rid FROM property_spatial_map WHERE property_id = ?`, prop.ID,
	).Scan(&rid); err != nil {
		t.Fatalf("spatial_map: %v", err)
	}

	var minLat float64
	if err := c.RawDB().QueryRowContext(ctx,
		`SELECT min_lat FROM property_rtree WHERE id = ?`, rid,
	).Scan(&minLat); err != nil {
		t.Fatalf("rtree: %v", err)
	}
	if math.Abs(minLat-hollywood[0]) > 1e-4 {
		t.Errorf("rtree not updated: min_lat = %v, want ~%v", minLat, hollywood[0])
	}
}

func TestRTreeRemovedOnDelete(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	prop := seedPropertyAt(t, c, "1 Delete Test", laCenter[0], laCenter[1])

	var rid int64
	if err := c.RawDB().QueryRowContext(ctx,
		`SELECT rid FROM property_spatial_map WHERE property_id = ?`, prop.ID,
	).Scan(&rid); err != nil {
		t.Fatalf("spatial_map before delete: %v", err)
	}

	if err := c.DeleteProperty(ctx, prop.ID); err != nil {
		t.Fatalf("DeleteProperty: %v", err)
	}

	var count int
	if err := c.RawDB().QueryRowContext(ctx,
		`SELECT COUNT(*) FROM property_rtree WHERE id = ?`, rid,
	).Scan(&count); err != nil {
		t.Fatalf("rtree count: %v", err)
	}
	if count != 0 {
		t.Errorf("rtree row still exists after property delete")
	}
}

// ---- FindPropertiesNearby ---------------------------------------------------

func TestFindPropertiesNearby_AllInRadius(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	seedPropertyAt(t, c, "1 Hollywood Blvd", hollywood[0], hollywood[1])       // ~8km
	seedPropertyAt(t, c, "2 Santa Monica Pier", santaMonica[0], santaMonica[1]) // ~23km
	seedPropertyAt(t, c, "3 Long Beach Port", longBeach[0], longBeach[1])       // ~32km

	// 50km radius should catch all three
	results, err := c.FindPropertiesNearby(ctx, laCenter[0], laCenter[1], 50)
	if err != nil {
		t.Fatalf("FindPropertiesNearby: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("len = %d, want 3", len(results))
	}
}

func TestFindPropertiesNearby_ExcludesFarProperties(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	seedPropertyAt(t, c, "1 Hollywood Blvd", hollywood[0], hollywood[1])  // ~8km — IN
	seedPropertyAt(t, c, "2 San Diego St", sanDiego[0], sanDiego[1])       // ~190km — OUT

	results, err := c.FindPropertiesNearby(ctx, laCenter[0], laCenter[1], 50)
	if err != nil {
		t.Fatalf("FindPropertiesNearby: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("len = %d, want 1 (san diego should be excluded)", len(results))
	}
	if results[0].Address != "1 Hollywood Blvd" {
		t.Errorf("Address = %q, want Hollywood", results[0].Address)
	}
}

func TestFindPropertiesNearby_SortedByDistance(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	// Insert in reverse distance order
	seedPropertyAt(t, c, "1 Long Beach", longBeach[0], longBeach[1])       // ~32km — farthest
	seedPropertyAt(t, c, "2 Santa Monica", santaMonica[0], santaMonica[1]) // ~23km
	seedPropertyAt(t, c, "3 Hollywood", hollywood[0], hollywood[1])        // ~8km — closest

	results, err := c.FindPropertiesNearby(ctx, laCenter[0], laCenter[1], 50)
	if err != nil {
		t.Fatalf("FindPropertiesNearby: %v", err)
	}
	if len(results) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(results))
	}
	for i := 1; i < len(results); i++ {
		if results[i].DistanceKm < results[i-1].DistanceKm {
			t.Errorf("results not sorted by distance: [%d]=%.2f > [%d]=%.2f",
				i-1, results[i-1].DistanceKm, i, results[i].DistanceKm)
		}
	}
}

func TestFindPropertiesNearby_DistanceIsAccurate(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	seedPropertyAt(t, c, "1 Hollywood Blvd", hollywood[0], hollywood[1])

	results, err := c.FindPropertiesNearby(ctx, laCenter[0], laCenter[1], 50)
	if err != nil {
		t.Fatalf("FindPropertiesNearby: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("no results")
	}
	expected := client.Haversine(laCenter[0], laCenter[1], hollywood[0], hollywood[1])
	if math.Abs(results[0].DistanceKm-expected) > 0.01 {
		t.Errorf("DistanceKm = %.4f, want %.4f", results[0].DistanceKm, expected)
	}
}

func TestFindPropertiesNearby_Empty(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	results, err := c.FindPropertiesNearby(ctx, laCenter[0], laCenter[1], 50)
	if err != nil {
		t.Fatalf("FindPropertiesNearby: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("len = %d, want 0 for empty db", len(results))
	}
}

// ---- FindListingsNearby -----------------------------------------------------

func TestFindListingsNearby_Basic(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	p1 := seedPropertyAt(t, c, "1 Hollywood Blvd", hollywood[0], hollywood[1])
	p2 := seedPropertyAt(t, c, "2 San Diego St", sanDiego[0], sanDiego[1])

	seedListing(t, c, p1.ID)
	seedListing(t, c, p2.ID)

	results, err := c.FindListingsNearby(ctx, laCenter[0], laCenter[1], 50, 0, 0)
	if err != nil {
		t.Fatalf("FindListingsNearby: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("len = %d, want 1 (san diego excluded)", len(results))
	}
}

func TestFindListingsNearby_PriceFilter(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	cheap := seedPropertyAt(t, c, "1 Cheap St", hollywood[0], hollywood[1])
	pricey := seedPropertyAt(t, c, "2 Pricey Ave", santaMonica[0], santaMonica[1])

	seedListing(t, c, cheap.ID, func(p *client.CreateListingParams) { p.ListPrice = 300000 })
	seedListing(t, c, pricey.ID, func(p *client.CreateListingParams) { p.ListPrice = 2000000 })

	results, err := c.FindListingsNearby(ctx, laCenter[0], laCenter[1], 50, 0, 500000)
	if err != nil {
		t.Fatalf("FindListingsNearby: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("len = %d, want 1 (only cheap listing)", len(results))
	}
	if results[0].ListPrice != 300000 {
		t.Errorf("ListPrice = %v, want 300000", results[0].ListPrice)
	}
}

// ---- FindRentalsNearby ------------------------------------------------------

func TestFindRentalsNearby_Basic(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	near := seedPropertyAt(t, c, "1 Near Rental", hollywood[0], hollywood[1])
	far := seedPropertyAt(t, c, "2 Far Rental", sanDiego[0], sanDiego[1])

	seedRental(t, c, near.ID)
	seedRental(t, c, far.ID)

	results, err := c.FindRentalsNearby(ctx, laCenter[0], laCenter[1], 50, 0, 0, 0)
	if err != nil {
		t.Fatalf("FindRentalsNearby: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("len = %d, want 1 (san diego excluded)", len(results))
	}
}

func TestFindRentalsNearby_RentFilter(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	p1 := seedPropertyAt(t, c, "1 Budget", hollywood[0], hollywood[1])
	p2 := seedPropertyAt(t, c, "2 Luxury", santaMonica[0], santaMonica[1])

	seedRental(t, c, p1.ID, func(p *client.CreateRentalParams) { p.MonthlyRent = 1500 })
	seedRental(t, c, p2.ID, func(p *client.CreateRentalParams) { p.MonthlyRent = 5000 })

	results, err := c.FindRentalsNearby(ctx, laCenter[0], laCenter[1], 50, 0, 2000, 0)
	if err != nil {
		t.Fatalf("FindRentalsNearby: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("len = %d, want 1 (only budget rental)", len(results))
	}
	if results[0].MonthlyRent != 1500 {
		t.Errorf("MonthlyRent = %v, want 1500", results[0].MonthlyRent)
	}
}

func TestFindRentalsNearby_BedroomFilter(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	p1 := seedPropertyAt(t, c, "1 Studio", hollywood[0], hollywood[1])
	p2 := seedPropertyAt(t, c, "2 TwoBed", santaMonica[0], santaMonica[1])

	seedRental(t, c, p1.ID, func(p *client.CreateRentalParams) { p.Bedrooms = 0 })
	seedRental(t, c, p2.ID, func(p *client.CreateRentalParams) { p.Bedrooms = 2 })

	results, err := c.FindRentalsNearby(ctx, laCenter[0], laCenter[1], 50, 0, 0, 2)
	if err != nil {
		t.Fatalf("FindRentalsNearby: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("len = %d, want 1 (2+ bed only)", len(results))
	}
	if results[0].Address != "2 TwoBed" {
		t.Errorf("Address = %q, want 2 TwoBed", results[0].Address)
	}
}

// ---- QueryBBox --------------------------------------------------------------

func TestQueryBBox_Direct(t *testing.T) {
	c, cleanup := newTestClient(t)
	defer cleanup()

	seedPropertyAt(t, c, "1 Inside", 34.05, -118.24)
	seedPropertyAt(t, c, "2 Outside", 35.00, -119.00)

	bbox := client.BBox{MinLat: 33.9, MaxLat: 34.2, MinLng: -118.5, MaxLng: -118.0}
	props, err := c.QueryBBox(ctx, bbox)
	if err != nil {
		t.Fatalf("QueryBBox: %v", err)
	}
	if len(props) != 1 {
		t.Errorf("len = %d, want 1", len(props))
	}
}
