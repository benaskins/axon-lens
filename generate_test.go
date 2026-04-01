package lens_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	lens "github.com/benaskins/axon-lens"
)

func TestFluxGenerator_GenerateImage(t *testing.T) {
	// Create a fake binary that writes a PNG to the --output path
	tmpDir := t.TempDir()
	fakeBin := filepath.Join(tmpDir, "fake-flux")

	// Create a real tiny PNG for the fake binary to copy
	pngData := createTestPNG(t, 64, 64)
	os.WriteFile(filepath.Join(tmpDir, "test.png"), pngData, 0644)

	// Script: find --output arg and copy the test PNG there
	script := `#!/bin/bash
output=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --output) output="$2"; shift 2 ;;
    *) shift ;;
  esac
done
cp "` + filepath.Join(tmpDir, "test.png") + `" "$output"
`
	os.WriteFile(fakeBin, []byte(script), 0755)

	gen := &lens.FluxGenerator{
		BinaryPath: fakeBin,
		Model:      "schnell",
		Width:      512,
		Height:     512,
		Steps:      4,
	}

	data, err := gen.GenerateImage(context.Background(), "a test image")
	if err != nil {
		t.Fatal(err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty image data")
	}

	// Should be valid PNG
	if len(data) < 8 || string(data[1:4]) != "PNG" {
		t.Error("output is not a valid PNG")
	}
}

func TestFluxGenerator_Defaults(t *testing.T) {
	// Create a fake binary that echoes its args
	tmpDir := t.TempDir()
	fakeBin := filepath.Join(tmpDir, "fake-flux")
	argsFile := filepath.Join(tmpDir, "args.txt")

	pngData := createTestPNG(t, 8, 8)
	os.WriteFile(filepath.Join(tmpDir, "test.png"), pngData, 0644)

	script := `#!/bin/bash
echo "$@" > "` + argsFile + `"
output=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --output) output="$2"; shift 2 ;;
    *) shift ;;
  esac
done
cp "` + filepath.Join(tmpDir, "test.png") + `" "$output"
`
	os.WriteFile(fakeBin, []byte(script), 0755)

	gen := &lens.FluxGenerator{BinaryPath: fakeBin}

	_, err := gen.GenerateImage(context.Background(), "test")
	if err != nil {
		t.Fatal(err)
	}

	args, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatal(err)
	}
	argsStr := string(args)

	// Check defaults were applied
	for _, expected := range []string{"--width 1024", "--height 1024", "--steps 4", "--model schnell"} {
		if !strings.Contains(argsStr, expected) {
			t.Errorf("expected args to contain %q, got %q", expected, argsStr)
		}
	}
}

func TestFluxGenerator_BinaryNotFound(t *testing.T) {
	gen := &lens.FluxGenerator{BinaryPath: "/nonexistent/binary"}

	_, err := gen.GenerateImage(context.Background(), "test")
	if err == nil {
		t.Error("expected error for missing binary")
	}
}
