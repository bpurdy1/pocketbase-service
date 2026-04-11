package rentalcomps

import (
	"github.com/pocketbase/pocketbase/core"
)

// RentalComp mirrors the rental_comps collection schema.
type RentalComp struct {
	Id           string   `json:"id"`
	Created      string   `json:"created"`
	Updated      string   `json:"updated"`
	Address      string   `json:"address"`
	Lat          float64  `json:"lat"`
	Lng          float64  `json:"lng"`
	Price        float64  `json:"price"`
	Sqft         float64  `json:"sqft"`
	PricePerSqft float64  `json:"price_per_sqft"`
	Building     string   `json:"building"`
	Website      string   `json:"website"`
	ListingURL   string   `json:"listing_url"`
	AddressURL   string   `json:"address_url"`
	DataSource   string   `json:"data_source"`
	Photos       []string `json:"photos"` // photo record IDs
}

// CreateOpts holds fields for creating a rental comp.
type CreateOpts struct {
	Address      string
	Lat          float64
	Lng          float64
	Price        float64
	Sqft         float64
	PricePerSqft float64
	Building     string
	Website      string
	ListingURL   string
	AddressURL   string
	DataSource   string
	Photos       []string // photo record IDs
}

// Client defines the interface for managing rental comps.
type Client interface {
	Create(opts CreateOpts) (*RentalComp, error)
	GetByID(id string) (*RentalComp, error)
	List() ([]*RentalComp, error)
	Delete(id string) error
}

type client struct {
	app core.App
}

// NewClient returns a new rental comps client.
func NewClient(app core.App) Client {
	return &client{app: app}
}

func fromRecord(r *core.Record) *RentalComp {
	return &RentalComp{
		Id:           r.Id,
		Created:      r.GetDateTime("created").String(),
		Updated:      r.GetDateTime("updated").String(),
		Address:      r.GetString("address"),
		Lat:          r.GetFloat("lat"),
		Lng:          r.GetFloat("lng"),
		Price:        r.GetFloat("price"),
		Sqft:         r.GetFloat("sqft"),
		PricePerSqft: r.GetFloat("price_per_sqft"),
		Building:     r.GetString("building"),
		Website:      r.GetString("website"),
		ListingURL:   r.GetString("listing_url"),
		AddressURL:   r.GetString("address_url"),
		DataSource:   r.GetString("data_source"),
		Photos:       r.GetStringSlice("photos"),
	}
}

func (c *client) Create(opts CreateOpts) (*RentalComp, error) {
	col, err := c.app.FindCollectionByNameOrId("rental_comps")
	if err != nil {
		return nil, err
	}

	record := core.NewRecord(col)
	record.Set("address", opts.Address)
	record.Set("lat", opts.Lat)
	record.Set("lng", opts.Lng)
	record.Set("price", opts.Price)
	record.Set("sqft", opts.Sqft)
	record.Set("price_per_sqft", opts.PricePerSqft)
	record.Set("building", opts.Building)
	record.Set("website", opts.Website)
	record.Set("listing_url", opts.ListingURL)
	record.Set("address_url", opts.AddressURL)
	record.Set("data_source", opts.DataSource)
	if len(opts.Photos) > 0 {
		record.Set("photos", opts.Photos)
	}

	if err := c.app.Save(record); err != nil {
		return nil, err
	}

	return fromRecord(record), nil
}

func (c *client) GetByID(id string) (*RentalComp, error) {
	record, err := c.app.FindRecordById("rental_comps", id)
	if err != nil {
		return nil, err
	}
	return fromRecord(record), nil
}

func (c *client) List() ([]*RentalComp, error) {
	records, err := c.app.FindRecordsByFilter(
		"rental_comps", "id != ''", "-created", 0, 0,
	)
	if err != nil {
		return nil, err
	}

	comps := make([]*RentalComp, len(records))
	for i, r := range records {
		comps[i] = fromRecord(r)
	}
	return comps, nil
}

func (c *client) Delete(id string) error {
	record, err := c.app.FindRecordById("rental_comps", id)
	if err != nil {
		return err
	}
	return c.app.Delete(record)
}
