// platformctl provides operator-facing administrative commands such as manifest
// validation and later smoke-test helpers. The CLI keeps operational behavior
// scriptable without pushing every task through the HTTP API.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/streanor/data-platform/backend/internal/config"
	"github.com/streanor/data-platform/backend/internal/manifests"
	"github.com/streanor/data-platform/backend/internal/orchestration"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: platformctl validate-manifests | remote [--server URL] [--token TOKEN] <command>")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "validate-manifests":
		if err := validateManifests(); err != nil {
			fmt.Fprintf(os.Stderr, "manifest validation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("manifest validation passed")
	case "remote":
		if err := runRemote(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "remote command failed: %v\n", err)
			os.Exit(1)
		}
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

func runRemote(args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	flagSet := flag.NewFlagSet("remote", flag.ContinueOnError)
	server := flagSet.String("server", cfg.APIBaseURL, "platform API base URL")
	token := flagSet.String("token", cfg.AdminToken, "admin token")
	if err := flagSet.Parse(args); err != nil {
		return err
	}

	command := strings.TrimSpace(strings.Join(flagSet.Args(), " "))
	if command == "" {
		return fmt.Errorf("remote command is required")
	}

	body, err := json.Marshal(map[string]string{"command": command})
	if err != nil {
		return err
	}

	request, err := http.NewRequest(http.MethodPost, strings.TrimRight(*server, "/")+"/api/v1/admin/terminal/execute", bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	if *token != "" {
		request.Header.Set("Authorization", "Bearer "+*token)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	payload, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("status %d: %s", response.StatusCode, strings.TrimSpace(string(payload)))
	}

	var result struct {
		Output []string `json:"output"`
	}
	if err := json.Unmarshal(payload, &result); err != nil {
		return fmt.Errorf("decode remote response: %w", err)
	}

	for _, line := range result.Output {
		fmt.Println(line)
	}
	return nil
}
