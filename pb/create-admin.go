package pb

import (
	"fmt"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// CreateAdmin creates a new superuser (admin) in PocketBase
func CreateAdmin(app *pocketbase.PocketBase, email, password string) (*core.Record, error) {
	collection, err := app.FindCollectionByNameOrId("_superusers")
	if err != nil {
		return nil, fmt.Errorf("failed to find _superusers collection: %w", err)
	}

	superuser := core.NewRecord(collection)
	superuser.Set("email", email)
	superuser.SetPassword(password)

	if err := app.Save(superuser); err != nil {
		return nil, fmt.Errorf("failed to create superuser: %w", err)
	}

	return superuser, nil
}

// AdminExists checks if an admin with the given email already exists
func AdminExists(app *pocketbase.PocketBase, email string) (bool, error) {
	record, err := app.FindAuthRecordByEmail("_superusers", email)
	if err != nil {
		return false, nil // Not found
	}
	return record != nil, nil
}

// EnsureAdmin creates an admin if one doesn't already exist with the given email
func EnsureAdmin(app *pocketbase.PocketBase, email, password string) (*core.Record, error) {
	exists, err := AdminExists(app, email)
	if err != nil {
		return nil, err
	}

	if exists {
		record, _ := app.FindAuthRecordByEmail("_superusers", email)
		return record, nil
	}

	return CreateAdmin(app, email, password)
}

// HasAnyAdmin checks if there are any real admin users (excluding the default installer)
func HasAnyAdmin(app *pocketbase.PocketBase) (bool, error) {
	superusers, err := app.FindAllRecords("_superusers")
	if err != nil {
		return false, err
	}

	for _, su := range superusers {
		email := su.GetString("email")
		if email != "" && email != "__pbinstaller@example.com" {
			return true, nil
		}
	}

	return false, nil
}
