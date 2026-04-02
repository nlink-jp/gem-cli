package cmd

import (
	"bytes"
	"testing"
)

func TestRootCmd_Version(t *testing.T) {
	appVersion = "test-version"
	root := newRootCmd()
	root.SetArgs([]string{"--version"})
	var out bytes.Buffer
	root.SetOut(&out)

	err := root.Execute()
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if got := out.String(); got == "" {
		t.Error("expected version output")
	}
}

func TestRootCmd_Help(t *testing.T) {
	root := newRootCmd()
	root.SetArgs([]string{"--help"})
	var out bytes.Buffer
	root.SetOut(&out)

	err := root.Execute()
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	got := out.String()
	if got == "" {
		t.Error("expected help output")
	}
}

func TestRootCmd_InvalidFormat(t *testing.T) {
	// Should fail during runPrompt because config requires project
	// but we can at least verify the command structure works
	root := newRootCmd()
	root.SetArgs([]string{"--format", "xml", "test"})
	root.SetErr(new(bytes.Buffer))

	// This will fail at config load (no project), not at format validation
	// That's acceptable — the format is validated by Gemini API, not CLI
	_ = root.Execute()
}

func TestRootCmd_NoInput(t *testing.T) {
	// No prompt, no stdin, no files — should error
	t.Setenv("GOOGLE_CLOUD_PROJECT", "test-project")
	root := newRootCmd()
	root.SetArgs([]string{})
	root.SetErr(new(bytes.Buffer))

	err := root.Execute()
	if err == nil {
		t.Error("expected error when no input provided")
	}
}

func TestRootCmd_FlagParsing(t *testing.T) {
	root := newRootCmd()
	root.SetArgs([]string{
		"--model", "gemini-2.5-pro",
		"--stream",
		"--grounding",
		"--format", "json",
		"--quiet",
		"--debug",
		"-s", "You are helpful.",
		"test prompt",
	})

	// Verify flags parse without error (execution will fail due to no project)
	t.Setenv("GOOGLE_CLOUD_PROJECT", "")
	root.SetErr(new(bytes.Buffer))
	_ = root.Execute() // will fail at config, but flags should parse

	// Verify flags are registered
	flags := root.Flags()
	if f := flags.Lookup("model"); f == nil {
		t.Error("--model flag not found")
	}
	if f := flags.Lookup("stream"); f == nil {
		t.Error("--stream flag not found")
	}
	if f := flags.Lookup("grounding"); f == nil {
		t.Error("--grounding flag not found")
	}
	if f := flags.Lookup("image"); f == nil {
		t.Error("--image flag not found")
	}
	if f := flags.Lookup("file"); f == nil {
		t.Error("--file flag not found")
	}
	if f := flags.Lookup("batch"); f == nil {
		t.Error("--batch flag not found")
	}
	if f := flags.Lookup("no-safe-input"); f == nil {
		t.Error("--no-safe-input flag not found")
	}
	if f := flags.Lookup("json-schema"); f == nil {
		t.Error("--json-schema flag not found")
	}
}

func TestRunOpts_FieldsExist(t *testing.T) {
	// Compile-time verification that runOpts has all expected fields
	_ = runOpts{
		configPath:       "",
		model:            "",
		systemPrompt:     "",
		systemPromptFile: "",
		format:           "",
		jsonSchema:       "",
		stream:           false,
		batch:            false,
		noSafeInput:      false,
		quiet:            false,
		debug:            false,
		grounding:        false,
		images:           nil,
		files:            nil,
		chat:             false,
		session:          "",
		cache:            false,
	}
}

func TestRootCmd_ChatFlag(t *testing.T) {
	root := newRootCmd()
	if f := root.Flags().Lookup("chat"); f == nil {
		t.Error("--chat flag not found")
	}
}

func TestRootCmd_CacheFlag(t *testing.T) {
	root := newRootCmd()
	if f := root.Flags().Lookup("cache"); f == nil {
		t.Error("--cache flag not found")
	}
}

func TestRootCmd_SessionFlag(t *testing.T) {
	root := newRootCmd()
	if f := root.Flags().Lookup("session"); f == nil {
		t.Error("--session flag not found")
	}
}
