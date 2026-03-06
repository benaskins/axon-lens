package lens

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ImageGenerator generates images from text prompts.
type ImageGenerator interface {
	GenerateImage(ctx context.Context, prompt string) ([]byte, error)
}

// FluxGenerator generates images by shelling out to a flux.swift binary.
type FluxGenerator struct {
	BinaryPath string
	Model      string // "schnell" or "dev"
	Width      int
	Height     int
	Steps      int
	Quantize   bool
}

// GenerateImage runs the flux.swift CLI and returns the resulting PNG bytes.
func (g *FluxGenerator) GenerateImage(ctx context.Context, prompt string) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "flux-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	outPath := filepath.Join(tmpDir, "output.png")

	width := g.Width
	if width == 0 {
		width = 1024
	}
	height := g.Height
	if height == 0 {
		height = 1024
	}
	steps := g.Steps
	if steps == 0 {
		steps = 4
	}
	model := g.Model
	if model == "" {
		model = "schnell"
	}

	args := []string{
		"--prompt", prompt,
		"--width", fmt.Sprintf("%d", width),
		"--height", fmt.Sprintf("%d", height),
		"--steps", fmt.Sprintf("%d", steps),
		"--model", model,
		"--output", outPath,
	}
	if g.Quantize {
		args = append(args, "--quantize")
	}

	cmd := exec.CommandContext(ctx, g.BinaryPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("flux generate: %w: %s", err, string(output))
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		return nil, fmt.Errorf("read generated image: %w", err)
	}

	return data, nil
}
