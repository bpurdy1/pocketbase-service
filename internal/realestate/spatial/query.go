package spatial

import (
	"context"
	"database/sql"
	"math"
	"sort"
)

// PropertyResult is a property row returned from a spatial query.
type PropertyResult struct {
	ID           string  `json:"id"`
	Organization string  `json:"organization"`
	PropertyName string  `json:"property_name"`
	Address      string  `json:"address"`
	City         string  `json:"city"`
	State        string  `json:"state"`
	ZipCode      string  `json:"zip_code"`
	County       string  `json:"county"`
	Lat          float64 `json:"lat"`
	Lng          float64 `json:"lng"`
	Price        float64 `json:"price"`
	Bedrooms     float64 `json:"bedrooms"`
	Bathrooms    float64 `json:"bathrooms"`
	Sqft         float64 `json:"sqft"`
	YearBuilt    float64 `json:"year_built"`
	Units        float64 `json:"number_of_units"`
	BuildingSF   float64 `json:"building_sf"`
	LotSF        float64 `json:"lot_sf"`
	PropertyType string  `json:"property_type"`
	Notes        string  `json:"notes"`
	Created      string  `json:"created"`
	Updated      string  `json:"updated"`
	DistanceKm   float64 `json:"distance_km"`
}

// BBox is a lat/lng bounding box.
type BBox struct {
	MinLat, MaxLat float64
	MinLng, MaxLng float64
}

// BBoxFromPoint returns a bounding box around a center point for a given radius in km.
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

// FindNearby returns all properties within radiusKm of (lat, lng) sorted by
// distance. Optionally scoped to a single organization (pass "" to skip).
//
// The R*Tree bbox pre-filters candidates; Haversine post-filters to the exact
// circle and computes the DistanceKm field.
func FindNearby(ctx context.Context, db *sql.DB, lat, lng, radiusKm float64, orgID string) ([]PropertyResult, error) {
	bbox := BBoxFromPoint(lat, lng, radiusKm)

	// Column order matches PocketBase's "properties" collection schema.
	q := `
		SELECT p.id, p.organization, p.property_name,
		       p.address, p.city, p.state, p.zip_code, p.county,
		       p.lat, p.lng,
		       p.price, p.bedrooms, p.bathrooms, p.sqft,
		       p.year_built, p.number_of_units, p.building_sf, p.lot_sf,
		       p.property_type, p.notes, p.created, p.updated
		FROM property_rtree r
		JOIN property_spatial_map m ON m.rid = r.id
		JOIN properties p           ON p.id  = m.property_id
		WHERE r.min_lat >= ? AND r.max_lat <= ?
		  AND r.min_lng >= ? AND r.max_lng <= ?`

	args := []any{bbox.MinLat, bbox.MaxLat, bbox.MinLng, bbox.MaxLng}
	if orgID != "" {
		q += ` AND p.organization = ?`
		args = append(args, orgID)
	}

	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []PropertyResult
	for rows.Next() {
		var p PropertyResult
		if err := rows.Scan(
			&p.ID, &p.Organization, &p.PropertyName,
			&p.Address, &p.City, &p.State, &p.ZipCode, &p.County,
			&p.Lat, &p.Lng,
			&p.Price, &p.Bedrooms, &p.Bathrooms, &p.Sqft,
			&p.YearBuilt, &p.Units, &p.BuildingSF, &p.LotSF,
			&p.PropertyType, &p.Notes, &p.Created, &p.Updated,
		); err != nil {
			return nil, err
		}
		dist := Haversine(lat, lng, p.Lat, p.Lng)
		if dist <= radiusKm {
			p.DistanceKm = dist
			results = append(results, p)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].DistanceKm < results[j].DistanceKm
	})
	return results, nil
}
