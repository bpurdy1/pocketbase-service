package client

import (
	"context"
	"database/sql"
	"math"
)

// BBox is a lat/lng bounding box.
type BBox struct {
	MinLat, MaxLat float64
	MinLng, MaxLng float64
}

// BBoxFromPoint computes a bounding box around a center point for a given radius in km.
func BBoxFromPoint(lat, lng, radiusKm float64) BBox {
	latDeg := radiusKm / 111.0
	lngDeg := radiusKm / (111.0 * math.Cos(lat*math.Pi/180))
	return BBox{
		MinLat: lat - latDeg,
		MaxLat: lat + latDeg,
		MinLng: lng - lngDeg,
		MaxLng: lng + lngDeg,
	}
}

// Haversine returns the great-circle distance in km between two lat/lng points.
func Haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// PropertyResult is a property row returned from a spatial query.
type PropertyResult struct {
	Property
	DistanceKm float64
}

// ListingResult is a sale listing + parent property from a spatial query.
type ListingResult struct {
	Property
	ListingID  string
	ListPrice  float64
	Status     string
	SourceUrl  string
	ExpiresAt  string
	MlsID      string
	DistanceKm float64
}

// RentalResult is a rental listing + parent property from a spatial query.
type RentalResult struct {
	Property
	RentalID      string
	MonthlyRent   float64
	Bedrooms      float64
	Bathrooms     float64
	Sqft          float64
	UnitNumber    string
	AvailableDate string
	SourceUrl     string
	ExpiresAt     string
	DistanceKm    float64
}

// FindPropertiesNearby returns all properties within radiusKm of the given point,
// sorted by distance ascending.
func (c *Client) FindPropertiesNearby(ctx context.Context, lat, lng, radiusKm float64) ([]PropertyResult, error) {
	bbox := BBoxFromPoint(lat, lng, radiusKm)
	const q = `
		SELECT p.id, p.address, p.city, p.state, p.zip_code, p.county,
		       p.lat, p.lng, p.property_name, p.property_type,
		       p.bedrooms, p.bathrooms, p.sqft, p.building_sf, p.lot_sf,
		       p.year_built, p.number_of_units, p.organization, p.notes,
		       p.created_at, p.updated_at
		FROM property_rtree r
		JOIN property_spatial_map m ON m.rid = r.id
		JOIN properties p ON p.id = m.property_id
		WHERE r.min_lat >= ? AND r.max_lat <= ?
		  AND r.min_lng >= ? AND r.max_lng <= ?`

	rows, err := c.db.QueryContext(ctx, q, bbox.MinLat, bbox.MaxLat, bbox.MinLng, bbox.MaxLng)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []PropertyResult
	for rows.Next() {
		var p Property
		if err := rows.Scan(
			&p.ID, &p.Address, &p.City, &p.State, &p.ZipCode, &p.County,
			&p.Lat, &p.Lng, &p.PropertyName, &p.PropertyType,
			&p.Bedrooms, &p.Bathrooms, &p.Sqft, &p.BuildingSf, &p.LotSf,
			&p.YearBuilt, &p.NumberOfUnits, &p.Organization, &p.Notes,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		dist := Haversine(lat, lng, p.Lat, p.Lng)
		if dist <= radiusKm {
			results = append(results, PropertyResult{Property: p, DistanceKm: dist})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	sortByDistance(results)
	return results, nil
}

// FindListingsNearby returns active for-sale listings within radiusKm,
// optionally filtered by min/max price (pass 0,0 to skip price filter).
func (c *Client) FindListingsNearby(ctx context.Context, lat, lng, radiusKm, minPrice, maxPrice float64) ([]ListingResult, error) {
	bbox := BBoxFromPoint(lat, lng, radiusKm)

	q := `
		SELECT p.id, p.address, p.city, p.state, p.zip_code, p.county,
		       p.lat, p.lng, p.property_name, p.property_type,
		       p.bedrooms, p.bathrooms, p.sqft, p.building_sf, p.lot_sf,
		       p.year_built, p.number_of_units, p.organization, p.notes,
		       p.created_at, p.updated_at,
		       pl.id, pl.list_price, pl.status, pl.source_url, pl.expires_at, pl.mls_id
		FROM listing_rtree r
		JOIN listing_spatial_map m ON m.rid = r.id
		JOIN property_listings pl ON pl.id = m.listing_id
		JOIN properties p ON p.id = pl.property_id
		WHERE r.min_lat >= ? AND r.max_lat <= ?
		  AND r.min_lng >= ? AND r.max_lng <= ?
		  AND pl.status = 'active'
		  AND pl.expires_at > datetime('now')`

	args := []any{bbox.MinLat, bbox.MaxLat, bbox.MinLng, bbox.MaxLng}
	if minPrice > 0 || maxPrice > 0 {
		q += ` AND pl.list_price >= ? AND pl.list_price <= ?`
		if maxPrice == 0 {
			maxPrice = math.MaxFloat64
		}
		args = append(args, minPrice, maxPrice)
	}
	q += ` ORDER BY pl.list_price ASC`

	rows, err := c.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ListingResult
	for rows.Next() {
		var r ListingResult
		if err := rows.Scan(
			&r.ID, &r.Address, &r.City, &r.State, &r.ZipCode, &r.County,
			&r.Lat, &r.Lng, &r.PropertyName, &r.PropertyType,
			&r.Bedrooms, &r.Bathrooms, &r.Sqft, &r.BuildingSf, &r.LotSf,
			&r.YearBuilt, &r.NumberOfUnits, &r.Organization, &r.Notes,
			&r.CreatedAt, &r.UpdatedAt,
			&r.ListingID, &r.ListPrice, &r.Status, &r.SourceUrl, &r.ExpiresAt, &r.MlsID,
		); err != nil {
			return nil, err
		}
		dist := Haversine(lat, lng, r.Lat, r.Lng)
		if dist <= radiusKm {
			r.DistanceKm = dist
			results = append(results, r)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// FindRentalsNearby returns active rental listings within radiusKm,
// optionally filtered by min/max rent and minimum bedrooms (pass 0 to skip).
func (c *Client) FindRentalsNearby(ctx context.Context, lat, lng, radiusKm, minRent, maxRent, minBedrooms float64) ([]RentalResult, error) {
	bbox := BBoxFromPoint(lat, lng, radiusKm)

	q := `
		SELECT p.id, p.address, p.city, p.state, p.zip_code, p.county,
		       p.lat, p.lng, p.property_name, p.property_type,
		       p.bedrooms, p.bathrooms, p.sqft, p.building_sf, p.lot_sf,
		       p.year_built, p.number_of_units, p.organization, p.notes,
		       p.created_at, p.updated_at,
		       rl.id, rl.monthly_rent, rl.bedrooms, rl.bathrooms, rl.sqft,
		       rl.unit_number, rl.available_date, rl.source_url, rl.expires_at
		FROM rental_rtree r
		JOIN rental_spatial_map m ON m.rid = r.id
		JOIN rental_listings rl ON rl.id = m.rental_id
		JOIN properties p ON p.id = rl.property_id
		WHERE r.min_lat >= ? AND r.max_lat <= ?
		  AND r.min_lng >= ? AND r.max_lng <= ?
		  AND rl.status = 'active'
		  AND rl.expires_at > datetime('now')`

	args := []any{bbox.MinLat, bbox.MaxLat, bbox.MinLng, bbox.MaxLng}
	if minRent > 0 || maxRent > 0 {
		q += ` AND rl.monthly_rent >= ? AND rl.monthly_rent <= ?`
		if maxRent == 0 {
			maxRent = math.MaxFloat64
		}
		args = append(args, minRent, maxRent)
	}
	if minBedrooms > 0 {
		q += ` AND rl.bedrooms >= ?`
		args = append(args, minBedrooms)
	}
	q += ` ORDER BY rl.monthly_rent ASC`

	rows, err := c.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []RentalResult
	for rows.Next() {
		var res RentalResult
		if err := rows.Scan(
			&res.ID, &res.Address, &res.City, &res.State, &res.ZipCode, &res.County,
			&res.Lat, &res.Lng, &res.PropertyName, &res.PropertyType,
			&res.Bedrooms, &res.Bathrooms, &res.Sqft, &res.BuildingSf, &res.LotSf,
			&res.YearBuilt, &res.NumberOfUnits, &res.Organization, &res.Notes,
			&res.CreatedAt, &res.UpdatedAt,
			&res.RentalID, &res.MonthlyRent, &res.Bedrooms, &res.Bathrooms, &res.Sqft,
			&res.UnitNumber, &res.AvailableDate, &res.SourceUrl, &res.ExpiresAt,
		); err != nil {
			return nil, err
		}
		dist := Haversine(lat, lng, res.Lat, res.Lng)
		if dist <= radiusKm {
			res.DistanceKm = dist
			results = append(results, res)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func sortByDistance(results []PropertyResult) {
	for i := 1; i < len(results); i++ {
		for j := i; j > 0 && results[j].DistanceKm < results[j-1].DistanceKm; j-- {
			results[j], results[j-1] = results[j-1], results[j]
		}
	}
}

// QueryBBox runs a raw bounding box query and returns matching properties.
// Use this when you already have a BBox computed externally.
func (c *Client) QueryBBox(ctx context.Context, bbox BBox) ([]Property, error) {
	const q = `
		SELECT p.id, p.address, p.city, p.state, p.zip_code, p.county,
		       p.lat, p.lng, p.property_name, p.property_type,
		       p.bedrooms, p.bathrooms, p.sqft, p.building_sf, p.lot_sf,
		       p.year_built, p.number_of_units, p.organization, p.notes,
		       p.created_at, p.updated_at
		FROM property_rtree r
		JOIN property_spatial_map m ON m.rid = r.id
		JOIN properties p ON p.id = m.property_id
		WHERE r.min_lat >= ? AND r.max_lat <= ?
		  AND r.min_lng >= ? AND r.max_lng <= ?`

	rows, err := c.db.QueryContext(ctx, q, bbox.MinLat, bbox.MaxLat, bbox.MinLng, bbox.MaxLng)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var props []Property
	for rows.Next() {
		var p Property
		if err := rows.Scan(
			&p.ID, &p.Address, &p.City, &p.State, &p.ZipCode, &p.County,
			&p.Lat, &p.Lng, &p.PropertyName, &p.PropertyType,
			&p.Bedrooms, &p.Bathrooms, &p.Sqft, &p.BuildingSf, &p.LotSf,
			&p.YearBuilt, &p.NumberOfUnits, &p.Organization, &p.Notes,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		props = append(props, p)
	}
	return props, rows.Err()
}

// ScanProperty is a helper for scanning a sql.Row into a Property.
func ScanProperty(row *sql.Row, dest *Property) error {
	return row.Scan(
		&dest.ID, &dest.Address, &dest.City, &dest.State, &dest.ZipCode, &dest.County,
		&dest.Lat, &dest.Lng, &dest.PropertyName, &dest.PropertyType,
		&dest.Bedrooms, &dest.Bathrooms, &dest.Sqft, &dest.BuildingSf, &dest.LotSf,
		&dest.YearBuilt, &dest.NumberOfUnits, &dest.Organization, &dest.Notes,
		&dest.CreatedAt, &dest.UpdatedAt,
	)
}
