package properties

import (
	"github.com/pocketbase/pocketbase/core"
)

// Property mirrors the properties collection schema.
type Property struct {
	Id           string  `json:"id"`
	Created      string  `json:"created"`
	Updated      string  `json:"updated"`
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
}

// CreateOpts holds the fields for creating a property.
type CreateOpts struct {
	Organization string
	PropertyName string
	Address      string
	City         string
	State        string
	ZipCode      string
	County       string
	Lat          float64
	Lng          float64
	Price        float64
	Bedrooms     float64
	Bathrooms    float64
	Sqft         float64
	YearBuilt    float64
	Units        float64
	BuildingSF   float64
	LotSF        float64
	PropertyType string
	Notes        string
}

// Client defines the interface for managing properties.
type Client interface {
	Create(opts CreateOpts) (*Property, error)
	GetByID(id string) (*Property, error)
	ListByOrg(orgID string) ([]*Property, error)
	Delete(id string) error
}

type client struct {
	app core.App
}

// NewClient returns a new properties client.
func NewClient(app core.App) Client {
	return &client{app: app}
}

func fromRecord(r *core.Record) *Property {
	return &Property{
		Id:           r.Id,
		Created:      r.GetDateTime("created").String(),
		Updated:      r.GetDateTime("updated").String(),
		Organization: r.GetString("organization"),
		PropertyName: r.GetString("property_name"),
		Address:      r.GetString("address"),
		City:         r.GetString("city"),
		State:        r.GetString("state"),
		ZipCode:      r.GetString("zip_code"),
		County:       r.GetString("county"),
		Lat:          r.GetFloat("lat"),
		Lng:          r.GetFloat("lng"),
		Price:        r.GetFloat("price"),
		Bedrooms:     r.GetFloat("bedrooms"),
		Bathrooms:    r.GetFloat("bathrooms"),
		Sqft:         r.GetFloat("sqft"),
		YearBuilt:    r.GetFloat("year_built"),
		Units:        r.GetFloat("number_of_units"),
		BuildingSF:   r.GetFloat("building_sf"),
		LotSF:        r.GetFloat("lot_sf"),
		PropertyType: r.GetString("property_type"),
		Notes:        r.GetString("notes"),
	}
}

func (c *client) Create(opts CreateOpts) (*Property, error) {
	col, err := c.app.FindCollectionByNameOrId("properties")
	if err != nil {
		return nil, err
	}

	record := core.NewRecord(col)
	record.Set("organization", opts.Organization)
	record.Set("property_name", opts.PropertyName)
	record.Set("address", opts.Address)
	record.Set("city", opts.City)
	record.Set("state", opts.State)
	record.Set("zip_code", opts.ZipCode)
	record.Set("county", opts.County)
	record.Set("lat", opts.Lat)
	record.Set("lng", opts.Lng)
	record.Set("price", opts.Price)
	record.Set("bedrooms", opts.Bedrooms)
	record.Set("bathrooms", opts.Bathrooms)
	record.Set("sqft", opts.Sqft)
	record.Set("year_built", opts.YearBuilt)
	record.Set("number_of_units", opts.Units)
	record.Set("building_sf", opts.BuildingSF)
	record.Set("lot_sf", opts.LotSF)
	record.Set("property_type", opts.PropertyType)
	record.Set("notes", opts.Notes)

	if err := c.app.Save(record); err != nil {
		return nil, err
	}

	return fromRecord(record), nil
}

func (c *client) GetByID(id string) (*Property, error) {
	record, err := c.app.FindRecordById("properties", id)
	if err != nil {
		return nil, err
	}
	return fromRecord(record), nil
}

func (c *client) ListByOrg(orgID string) ([]*Property, error) {
	records, err := c.app.FindRecordsByFilter(
		"properties",
		"organization = {:orgId}",
		"-created", 0, 0,
		map[string]any{"orgId": orgID},
	)
	if err != nil {
		return nil, err
	}

	props := make([]*Property, len(records))
	for i, r := range records {
		props[i] = fromRecord(r)
	}
	return props, nil
}

func (c *client) Delete(id string) error {
	record, err := c.app.FindRecordById("properties", id)
	if err != nil {
		return err
	}
	return c.app.Delete(record)
}
