// platformctl provides operator-facing administrative commands such as manifest
// validation and later smoke-test helpers. The CLI keeps operational behavior
// scriptable without pushing every task through the HTTP API.
package main

import (
	"fmt"
	"os"

	"github.com/streanor/data-platform/backend/internal/config"
	"github.com/streanor/data-platform/backend/internal/manifests"
	"github.com/streanor/data-platform/backend/internal/orchestration"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: platformctl validate-manifests")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "validate-manifests":
		if err := validateManifests(); err != nil {
			fmt.Fprintf(os.Stderr, "manifest validation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("manifest validation passed")
	default:
		fmt.Printf("unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func validateManifests() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	loader := manifests.NewLoader(cfg.ManifestRoot)
	pipelines, err := loader.LoadPipelines()
	if err != nil {
		return err
	}

	for _, pipeline := range pipelines {
		if err := orchestration.ValidatePipeline(pipeline); err != nil {
			return fmt.Errorf("pipeline %s: %w", pipeline.ID, err)
		}
	}

	_, err = loader.LoadAssets()
	return err
}
