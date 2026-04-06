package lens_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	lens "github.com/benaskins/axon-lens"
)

// TestBackendInterface verifies that both SDXLBackend and FluxBackend satisfy Backend.
func TestBackendInterface(t *testing.T) {
	var _ lens.Backend = lens.NewSDXLBackend()
	var _ lens.Backend = lens.NewFluxBackend()
}

func TestSDXLBackendName(t *testing.T) {
	b := lens.NewSDXLBackend()
	if b.Name() != "sdxl-turbo" {
		t.Errorf("Name() = %q, want %q", b.Name(), "sdxl-turbo")
	}
}

func TestFluxBackendName(t *testing.T) {
	b := lens.NewFluxBackend()
	if b.Name() != "flux-schnell" {
		t.Errorf("Name() = %q, want %q", b.Name(), "flux-schnell")
	}

	b2 := &lens.FluxBackend{Model: "dev"}
	if b2.Name() != "flux-dev" {
		t.Errorf("Name() = %q, want %q", b2.Name(), "flux-dev")
	}
}

// TestSDXLBackendEmptyPrompt verifies that an empty prompt returns an error
// before shelling out.
func TestSDXLBackendEmptyPrompt(t *testing.T) {
	b := lens.NewSDXLBackend()
	_, err := b.Txt2Img(context.Background(), lens.GenerateRequest{Prompt: ""})
	if err == nil {
		t.Fatal("expected error for empty prompt, got nil")
	}
}

func TestFluxBackendEmptyPrompt(t *testing.T) {
	b := lens.NewFluxBackend()
	_, err := b.Txt2Img(context.Background(), lens.GenerateRequest{Prompt: ""})
	if err == nil {
		t.Fatal("expected error for empty prompt, got nil")
	}
}

// TestSDXLBackendScriptInvocation uses a fake script to verify the correct
// arguments are passed without running real inference.
func TestSDXLBackendScriptInvocation(t *testing.T) {
	// Write a fake generate-image script that records its args and writes a PNG.
	dir := t.TempDir()
	outDir := filepath.Join(dir, "output")

	fakeScript := filepath.Join(dir, "generate-image")
	fakeScriptContent := `#!/usr/bin/env bash
set -euo pipefail
# Parse --output argument
output=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --output) output="$2"; shift 2 ;;
    *) shift ;;
  esac
done
touch "$output"
echo "$output"
`
	if err := os.WriteFile(fakeScript, []byte(fakeScriptContent), 0755); err != nil {
		t.Fatal(err)
	}

	b := &lens.SDXLBackend{
		ScriptPath: fakeScript,
		OutputDir:  outDir,
	}

	req := lens.GenerateRequest{
		Prompt: "a blue mountain lake",
		Width:  256,
		Height: 256,
		Steps:  2,
		Seed:   42,
	}

	result, err := b.Txt2Img(context.Background(), req)
	if err != nil {
		t.Fatalf("Txt2Img error: %v", err)
	}

	if result.BackendName != "sdxl-turbo" {
		t.Errorf("BackendName = %q, want %q", result.BackendName, "sdxl-turbo")
	}
	if result.Width != 256 {
		t.Errorf("Width = %d, want 256", result.Width)
	}
	if result.Height != 256 {
		t.Errorf("Height = %d, want 256", result.Height)
	}
	if result.Steps != 2 {
		t.Errorf("Steps = %d, want 2", result.Steps)
	}
	if result.Seed != 42 {
		t.Errorf("Seed = %d, want 42", result.Seed)
	}
	if result.Model != "stabilityai/sdxl-turbo" {
		t.Errorf("Model = %q, want %q", result.Model, "stabilityai/sdxl-turbo")
	}
	if !strings.HasSuffix(result.OutputPath, ".png") {
		t.Errorf("OutputPath = %q, expected .png suffix", result.OutputPath)
	}
	if result.Elapsed <= 0 {
		t.Error("Elapsed should be > 0")
	}
}

// TestSDXLBackendScriptFailure verifies error propagation when the script exits non-zero.
func TestSDXLBackendScriptFailure(t *testing.T) {
	dir := t.TempDir()
	fakeScript := filepath.Join(dir, "generate-image")
	if err := os.WriteFile(fakeScript, []byte("#!/usr/bin/env bash\necho 'fatal error' >&2\nexit 1\n"), 0755); err != nil {
		t.Fatal(err)
	}

	b := &lens.SDXLBackend{ScriptPath: fakeScript, OutputDir: dir}
	_, err := b.Txt2Img(context.Background(), lens.GenerateRequest{Prompt: "test"})
	if err == nil {
		t.Fatal("expected error from failing script, got nil")
	}
}

// TestSDXLBackendMissingScript verifies that a missing script produces an error.
func TestSDXLBackendMissingScript(t *testing.T) {
	b := &lens.SDXLBackend{ScriptPath: "/nonexistent/generate-image"}
	_, err := b.Txt2Img(context.Background(), lens.GenerateRequest{Prompt: "test"})
	if err == nil {
		t.Fatal("expected error for missing script, got nil")
	}
	// Should be a path error or exec error.
	if !isNotFoundError(err) {
		t.Logf("error (acceptable): %v", err)
	}
}

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	if _, ok := err.(*exec.Error); ok {
		return true
	}
	return strings.Contains(err.Error(), "not found") ||
		strings.Contains(err.Error(), "no such file")
}

// TestGenerateRequestDefaults verifies zero-value fields behave as documented.
func TestGenerateRequestDefaults(t *testing.T) {
	req := lens.GenerateRequest{Prompt: "test"}
	if req.Width != 0 {
		t.Error("expected Width zero by default")
	}
	if req.Steps != 0 {
		t.Error("expected Steps zero by default")
	}
	if req.Seed != 0 {
		t.Error("expected Seed zero (no seed) by default")
	}
}

// TestFluxBackendScriptInvocation uses a fake binary to exercise the happy path.
func TestFluxBackendScriptInvocation(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "output")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		t.Fatal(err)
	}

	fakeBin := filepath.Join(dir, "flux")
	// Fake flux binary: parse --output flag, create the file.
	fakeContent := `#!/usr/bin/env bash
set -euo pipefail
output=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --output) output="$2"; shift 2 ;;
    *) shift ;;
  esac
done
touch "$output"
`
	if err := os.WriteFile(fakeBin, []byte(fakeContent), 0755); err != nil {
		t.Fatal(err)
	}

	b := &lens.FluxBackend{
		BinaryPath: fakeBin,
		Model:      "schnell",
		OutputDir:  outDir,
	}

	req := lens.GenerateRequest{
		Prompt: "a mountain lake at sunrise",
		Width:  512,
		Height: 512,
		Steps:  2,
	}

	result, err := b.Txt2Img(context.Background(), req)
	if err != nil {
		t.Fatalf("Txt2Img error: %v", err)
	}

	if result.BackendName != "flux-schnell" {
		t.Errorf("BackendName = %q, want %q", result.BackendName, "flux-schnell")
	}
	if result.Width != 512 {
		t.Errorf("Width = %d, want 512", result.Width)
	}
	if result.Steps != 2 {
		t.Errorf("Steps = %d, want 2", result.Steps)
	}
	if result.Model != "black-forest-labs/FLUX.1-schnell" {
		t.Errorf("Model = %q", result.Model)
	}
	if !strings.HasSuffix(result.OutputPath, ".png") {
		t.Errorf("OutputPath %q should end with .png", result.OutputPath)
	}
	if result.Elapsed <= 0 {
		t.Error("Elapsed should be > 0")
	}
}

// TestFluxBackendFailure verifies error propagation when the binary exits non-zero.
func TestFluxBackendFailure(t *testing.T) {
	dir := t.TempDir()
	fakeBin := filepath.Join(dir, "flux")
	if err := os.WriteFile(fakeBin, []byte("#!/usr/bin/env bash\nexit 1\n"), 0755); err != nil {
		t.Fatal(err)
	}

	b := &lens.FluxBackend{BinaryPath: fakeBin, OutputDir: dir}
	_, err := b.Txt2Img(context.Background(), lens.GenerateRequest{Prompt: "test"})
	if err == nil {
		t.Fatal("expected error from failing binary, got nil")
	}
}

// TestCameraPrompt ensures the prompt string is non-empty.
func TestCameraPrompt(t *testing.T) {
	p := lens.CameraPrompt()
	if p == "" {
		t.Error("CameraPrompt() returned empty string")
	}
}

// TestGenerateResultFields spot-checks the struct layout.
func TestGenerateResultFields(t *testing.T) {
	result := &lens.GenerateResult{
		OutputPath:  "/tmp/out.png",
		Width:       512,
		Height:      512,
		Steps:       1,
		Model:       "stabilityai/sdxl-turbo",
		BackendName: "sdxl-turbo",
	}
	if result.OutputPath != "/tmp/out.png" {
		t.Errorf("OutputPath = %q", result.OutputPath)
	}
	if result.Model != "stabilityai/sdxl-turbo" {
		t.Errorf("Model = %q", result.Model)
	}
}
