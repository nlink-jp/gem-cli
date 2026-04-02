package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("GOOGLE_CLOUD_PROJECT", "test-project")
	// No config file, no other env vars — should use defaults
	cfg, err := Load("/nonexistent/path.toml")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.GCP.Project != "test-project" {
		t.Errorf("project = %q, want test-project", cfg.GCP.Project)
	}
	if cfg.GCP.Location != "us-central1" {
		t.Errorf("location = %q, want us-central1", cfg.GCP.Location)
	}
	if cfg.Model.Name != "gemini-2.5-flash" {
		t.Errorf("model = %q, want gemini-2.5-flash", cfg.Model.Name)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	t.Setenv("GEM_CLI_PROJECT", "env-project")
	t.Setenv("GEM_CLI_LOCATION", "asia-northeast1")
	t.Setenv("GEM_CLI_MODEL", "gemini-2.5-pro")

	cfg, err := Load("/nonexistent/path.toml")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.GCP.Project != "env-project" {
		t.Errorf("project = %q, want env-project", cfg.GCP.Project)
	}
	if cfg.GCP.Location != "asia-northeast1" {
		t.Errorf("location = %q, want asia-northeast1", cfg.GCP.Location)
	}
	if cfg.Model.Name != "gemini-2.5-pro" {
		t.Errorf("model = %q, want gemini-2.5-pro", cfg.Model.Name)
	}
}

func TestLoad_TOMLFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")
	content := `[gcp]
project  = "toml-project"
location = "europe-west1"

[model]
name = "gemini-2.0-flash"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	// Clear env to ensure TOML takes effect
	t.Setenv("GOOGLE_CLOUD_PROJECT", "")
	t.Setenv("GEM_CLI_PROJECT", "")

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.GCP.Project != "toml-project" {
		t.Errorf("project = %q, want toml-project", cfg.GCP.Project)
	}
	if cfg.GCP.Location != "europe-west1" {
		t.Errorf("location = %q, want europe-west1", cfg.GCP.Location)
	}
	if cfg.Model.Name != "gemini-2.0-flash" {
		t.Errorf("model = %q, want gemini-2.0-flash", cfg.Model.Name)
	}
}

func TestLoad_EnvOverridesToml(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")
	content := `[gcp]
project = "toml-project"
[model]
name = "toml-model"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	// Clear env vars that would override TOML values
	t.Setenv("GOOGLE_CLOUD_PROJECT", "")
	t.Setenv("GEM_CLI_PROJECT", "")
	t.Setenv("GEM_CLI_MODEL", "env-model")

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	// Env should override TOML
	if cfg.Model.Name != "env-model" {
		t.Errorf("model = %q, want env-model (env override)", cfg.Model.Name)
	}
	// TOML should still apply for non-overridden fields
	if cfg.GCP.Project != "toml-project" {
		t.Errorf("project = %q, want toml-project", cfg.GCP.Project)
	}
}

func TestLoad_MissingProject(t *testing.T) {
	t.Setenv("GOOGLE_CLOUD_PROJECT", "")
	t.Setenv("GEM_CLI_PROJECT", "")

	_, err := Load("/nonexistent/path.toml")
	if err == nil {
		t.Error("expected error when project is not set")
	}
}
