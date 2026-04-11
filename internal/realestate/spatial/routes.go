package spatial

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/pocketbase/pocketbase/core"
)

func registerRoutes(app core.App, db *sql.DB) {
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		// GET /api/spatial/properties/nearby
		//
		// Query params:
		//   lat      float64  (required) center latitude
		//   lng      float64  (required) center longitude
		//   radius   float64  (optional, default 25) search radius in km
		//   org      string   (optional) filter to a specific organization ID
		//
		// Requires authentication. Returns properties sorted by distance ascending.
		e.Router.GET("/api/spatial/properties/nearby", func(re *core.RequestEvent) error {
			if re.Auth == nil {
				return re.JSON(http.StatusUnauthorized, map[string]string{"error": "authentication required"})
			}

			lat, err := parseFloat(re.Request.URL.Query().Get("lat"))
			if err != nil {
				return re.JSON(http.StatusBadRequest, map[string]string{"error": "lat is required"})
			}
			lng, err := parseFloat(re.Request.URL.Query().Get("lng"))
			if err != nil {
				return re.JSON(http.StatusBadRequest, map[string]string{"error": "lng is required"})
			}

			radius := 25.0
			if r := re.Request.URL.Query().Get("radius"); r != "" {
				if v, err := strconv.ParseFloat(r, 64); err == nil && v > 0 {
					radius = v
				}
			}

			orgID := re.Request.URL.Query().Get("org")

			results, err := FindNearby(re.Request.Context(), db, lat, lng, radius, orgID)
			if err != nil {
				app.Logger().Error("spatial/nearby query failed", "error", err)
				return re.JSON(http.StatusInternalServerError, map[string]string{"error": "query failed"})
			}

			return re.JSON(http.StatusOK, map[string]any{
				"items":      results,
				"totalItems": len(results),
				"lat":        lat,
				"lng":        lng,
				"radiusKm":   radius,
			})
		})

		return e.Next()
	})
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
