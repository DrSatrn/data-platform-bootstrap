// This file owns manifest-to-store projection for the metadata catalog. The
// goal is to keep synchronization explicit and lifecycle-driven rather than
// hiding write behavior inside read handlers.
package metadata

import "fmt"

// ProjectStore loads the current manifest-backed assets and writes them into
// the configured metadata store.
func ProjectStore(loader AssetLoader, store Store) error {
	if loader == nil || store == nil {
		return nil
	}
	assets, err := loader.LoadAssets()
	if err != nil {
		return fmt.Errorf("load assets for projection: %w", err)
	}
	if err := store.SyncAssets(assets); err != nil {
		return fmt.Errorf("sync metadata store: %w", err)
	}
	return nil
}
