package config

import "testing"

func TestLoadParsesAlertWebhookSettings(t *testing.T) {
	t.Setenv("PLATFORM_ALERT_RUN_FAILURE_WEBHOOK_URLS", "https://example.invalid/run-a, https://example.invalid/run-b")
	t.Setenv("PLATFORM_ALERT_ASSET_WARNING_WEBHOOK_URLS", "https://example.invalid/asset-a")
	t.Setenv("PLATFORM_ALERT_WEBHOOK_TIMEOUT", "9s")

	settings, err := Load()
	if err != nil {
		t.Fatalf("load settings: %v", err)
	}
	if len(settings.RunFailureWebhookURLs) != 2 {
		t.Fatalf("expected 2 run failure webhook URLs, got %d", len(settings.RunFailureWebhookURLs))
	}
	if settings.RunFailureWebhookURLs[1] != "https://example.invalid/run-b" {
		t.Fatalf("unexpected run failure URLs: %#v", settings.RunFailureWebhookURLs)
	}
	if len(settings.AssetWarningWebhookURLs) != 1 || settings.AssetWarningWebhookURLs[0] != "https://example.invalid/asset-a" {
		t.Fatalf("unexpected asset warning URLs: %#v", settings.AssetWarningWebhookURLs)
	}
	if settings.AlertWebhookTimeout.String() != "9s" {
		t.Fatalf("expected 9s alert timeout, got %s", settings.AlertWebhookTimeout)
	}
}
