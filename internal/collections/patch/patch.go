package patch

import (
	"github.com/pocketbase/pocketbase/core"
)

// Func is a function that mutates a collection and returns true if it changed.
type Func func(col *core.Collection) bool

// Collection finds an existing collection and applies patches to it.
// Only saves if at least one patch reports a change.
// Returns nil if the collection does not exist.
func Collection(app core.App, name string, patches ...Func) error {
	col, err := app.FindCollectionByNameOrId(name)
	if err != nil {
		return nil // collection doesn't exist yet, nothing to patch
	}

	changed := false
	for _, p := range patches {
		if p(col) {
			changed = true
		}
	}

	if changed {
		return app.Save(col)
	}
	return nil
}

// Field adds a field if it doesn't already exist.
func Field(field core.Field) Func {
	return func(col *core.Collection) bool {
		if col.Fields.GetByName(field.GetName()) != nil {
			return false
		}
		col.Fields.Add(field)
		return true
	}
}

// AutodateFields adds created/updated autodate fields if missing.
func AutodateFields() Func {
	return func(col *core.Collection) bool {
		changed := false
		if col.Fields.GetByName("created") == nil {
			col.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
			changed = true
		}
		if col.Fields.GetByName("updated") == nil {
			col.Fields.Add(&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true})
			changed = true
		}
		return changed
	}
}

// Index adds a named index if it doesn't already exist.
func Index(name string, unique bool, columns string) Func {
	return func(col *core.Collection) bool {
		if col.GetIndex(name) != "" {
			return false
		}
		col.AddIndex(name, unique, columns, "")
		return true
	}
}

// RelationField mutates an existing RelationField in place.
func RelationField(name string, mutate func(f *core.RelationField) bool) Func {
	return func(col *core.Collection) bool {
		f, ok := col.Fields.GetByName(name).(*core.RelationField)
		if !ok {
			return false
		}
		return mutate(f)
	}
}

// TextField mutates an existing TextField in place.
// The mutate func receives the field and returns true if it changed.
func TextField(name string, mutate func(f *core.TextField) bool) Func {
	return func(col *core.Collection) bool {
		f, ok := col.Fields.GetByName(name).(*core.TextField)
		if !ok {
			return false
		}
		return mutate(f)
	}
}

// ClearRules sets all access rules to nil so they will be re-applied by tenancy on next startup.
func ClearRules() Func {
	return func(col *core.Collection) bool {
		if col.ListRule == nil && col.ViewRule == nil &&
			col.CreateRule == nil && col.UpdateRule == nil && col.DeleteRule == nil {
			return false
		}
		col.ListRule = nil
		col.ViewRule = nil
		col.CreateRule = nil
		col.UpdateRule = nil
		col.DeleteRule = nil
		return true
	}
}
