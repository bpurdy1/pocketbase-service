package client

import (
	"fmt"
	"time"

	"sqlite-realestate/client/db"
)

// --- Core domain types -------------------------------------------------------
// These are wrapper structs (not aliases) so we can attach methods.
// sqlc returns db.* types; use the wrap helpers or Client methods below.

type Property struct{ db.Property }

func (p Property) FullAddress() string {
	return fmt.Sprintf("%s, %s, %s %s", p.Address, p.City, p.State, p.ZipCode)
}

func (p Property) HasCoords() bool {
	return p.Lat != 0 || p.Lng != 0
}

type PropertyListing struct{ db.PropertyListing }

func (l PropertyListing) IsExpired() bool {
	t, err := time.Parse(time.RFC3339, l.ExpiresAt)
	if err != nil {
		return false
	}
	return time.Now().After(t)
}

func (l PropertyListing) IsActive() bool {
	return l.Status == "active" && !l.IsExpired()
}

type RentalListing struct{ db.RentalListing }

func (r RentalListing) IsExpired() bool {
	t, err := time.Parse(time.RFC3339, r.ExpiresAt)
	if err != nil {
		return false
	}
	return time.Now().After(t)
}

func (r RentalListing) IsActive() bool {
	return r.Status == "active" && !r.IsExpired()
}

func (r RentalListing) RentPerBedroom() float64 {
	if r.Bedrooms == 0 {
		return r.MonthlyRent
	}
	return r.MonthlyRent / r.Bedrooms
}

// --- Wrap helpers ------------------------------------------------------------
// Convert db.* → client.* when you have a raw sqlc result and want methods.

func WrapProperty(p db.Property) Property              { return Property{p} }
func WrapListing(l db.PropertyListing) PropertyListing { return PropertyListing{l} }
func WrapRental(r db.RentalListing) RentalListing      { return RentalListing{r} }

func WrapProperties(ps []db.Property) []Property {
	out := make([]Property, len(ps))
	for i, p := range ps {
		out[i] = WrapProperty(p)
	}
	return out
}

func WrapListings(ls []db.PropertyListing) []PropertyListing {
	out := make([]PropertyListing, len(ls))
	for i, l := range ls {
		out[i] = WrapListing(l)
	}
	return out
}

func WrapRentals(rs []db.RentalListing) []RentalListing {
	out := make([]RentalListing, len(rs))
	for i, r := range rs {
		out[i] = WrapRental(r)
	}
	return out
}

// --- Pass-through types (aliases — no methods needed) ------------------------

type PropertiesHistory = db.PropertiesHistory
type PropertyListingsHistory = db.PropertyListingsHistory
type RentalListingsHistory = db.RentalListingsHistory
type PropertyPhoto = db.PropertyPhoto
type ListingSource = db.ListingSource
type PropertySpatialMap = db.PropertySpatialMap
type PropertyRtree = db.PropertyRtree
type ListingSpatialMap = db.ListingSpatialMap
type ListingRtree = db.ListingRtree
type RentalSpatialMap = db.RentalSpatialMap
type RentalRtree = db.RentalRtree

// --- Params (aliases — no methods needed) ------------------------------------

type CreatePropertyParams = db.CreatePropertyParams
type UpdatePropertyParams = db.UpdatePropertyParams
type UpdatePropertyCoordsParams = db.UpdatePropertyCoordsParams
type ListPropertiesParams = db.ListPropertiesParams
type ListPropertiesByCityStateParams = db.ListPropertiesByCityStateParams
type ListPropertiesByOrgParams = db.ListPropertiesByOrgParams
type ListPropertiesByTypeParams = db.ListPropertiesByTypeParams

type CreateListingParams = db.CreateListingParams
type UpdateListingParams = db.UpdateListingParams
type UpdateListingStatusParams = db.UpdateListingStatusParams
type UpdateListingTTLParams = db.UpdateListingTTLParams
type ListActiveListingsParams = db.ListActiveListingsParams
type ListAllListingHistoryParams = db.ListAllListingHistoryParams
type GetListingHistoryParams = db.GetListingHistoryParams
type ArchiveListingParams = db.ArchiveListingParams

type CreateRentalParams = db.CreateRentalParams
type UpdateRentalParams = db.UpdateRentalParams
type UpdateRentalStatusParams = db.UpdateRentalStatusParams
type UpdateRentalTTLParams = db.UpdateRentalTTLParams
type GetActiveRentalByPropertyUnitParams = db.GetActiveRentalByPropertyUnitParams
type ListActiveRentalsParams = db.ListActiveRentalsParams
type ListActiveRentalsByPriceRangeParams = db.ListActiveRentalsByPriceRangeParams
type ListActiveRentalsByBedroomsParams = db.ListActiveRentalsByBedroomsParams
type GetRentalHistoryParams = db.GetRentalHistoryParams
type ArchiveRentalParams = db.ArchiveRentalParams

type GetPropertyHistoryParams = db.GetPropertyHistoryParams
type GetPropertyHistoryByChangeTypeParams = db.GetPropertyHistoryByChangeTypeParams
type ListAllPropertyHistoryParams = db.ListAllPropertyHistoryParams

type AddPhotoParams = db.AddPhotoParams
type UpsertListingSourceParams = db.UpsertListingSourceParams
type ListRecentlySeenSourcesParams = db.ListRecentlySeenSourcesParams

// --- Row types ---------------------------------------------------------------

type GetPropertyWithListingAndRentalRow = db.GetPropertyWithListingAndRentalRow
type ListActiveListingsRow = db.ListActiveListingsRow
type ListActiveRentalsRow = db.ListActiveRentalsRow
type ListActiveRentalsByPriceRangeRow = db.ListActiveRentalsByPriceRangeRow
type ListActiveRentalsByBedroomsRow = db.ListActiveRentalsByBedroomsRow
type ListAllListingHistoryRow = db.ListAllListingHistoryRow
type ListRecentlySeenSourcesRow = db.ListRecentlySeenSourcesRow
