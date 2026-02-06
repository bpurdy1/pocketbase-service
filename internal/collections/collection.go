package collections

import (
	"fmt"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type CollectionManager struct {
	app *pocketbase.PocketBase

	// Cached collections (lazy loaded)
	user   *core.Collection
	config *core.Collection
}

func NewCollectionManager(app *pocketbase.PocketBase) *CollectionManager {
	// Don't query DB here - it's not ready yet
	return &CollectionManager{
		app: app,
	}
}

// Users returns the users collection (lazy loaded)
func (cm *CollectionManager) Users() (*core.Collection, error) {
	if cm.user == nil {
		c, err := cm.app.FindCollectionByNameOrId("users")
		if err != nil {
			return nil, fmt.Errorf("users collection not found: %w", err)
		}
		cm.user = c
	}
	return cm.user, nil
}

// Config returns the config collection (lazy loaded)
func (cm *CollectionManager) Config() (*core.Collection, error) {
	if cm.config == nil {
		c, err := cm.app.FindCollectionByNameOrId("config")
		if err != nil {
			return nil, fmt.Errorf("config collection not found: %w", err)
		}
		cm.config = c
	}
	return cm.config, nil
}
