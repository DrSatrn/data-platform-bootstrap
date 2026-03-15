// This file owns manifest-to-store projection for the metadata catalog. The
// goal is to keep synchronization explicit and lifecycle-driven rather than
// hiding write behavior inside read handlers.
package metadata

import "fmt"

// ProjectStore loads the current manifest-backed assets and seeds them into the
// configured metadata store. The store keeps database-managed annotations as
// the runtime truth, so seeding only refreshes the manifest-backed structural
// fields rather than replacing operator edits.
func ProjectStore(loader AssetLoader, store Store) error {
	if loader == nil || store == nil {
		return nil
	}
	assets, err := loader.LoadAssets()
	if err != nil {
		return fmt.Errorf("load assets for projection: %w", err)
	}
	if err := store.SeedAssets(assets); err != nil {
		return fmt.Errorf("seed metadata store: %w", err)
	}
	return nil
}
