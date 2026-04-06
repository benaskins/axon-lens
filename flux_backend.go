package lens

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// FluxBackend generates images using the flux.swift CLI (FLUX.1 Schnell or Dev).
// It implements Backend and is the typed successor to FluxGenerator.
//
// Install flux.swift via: just install-flux  (in the lamina workspace root)
type FluxBackend struct {
	// BinaryPath is the path to the flux binary.
	// Defaults to "flux" (resolved via PATH).
	BinaryPath string

	// Model is "schnell" (default) or "dev".
	Model string

	// Quantize enables 4-bit quantization to reduce memory usage.
	Quantize bool

	// OutputDir is the default directory for generated images.
	// Defaults to ~/generated-images.
	OutputDir string
}

// NewFluxBackend returns a FluxBackend using the flux binary from PATH.
func NewFluxBackend() *FluxBackend {
	return &FluxBackend{BinaryPath: "flux", Model: "schnell"}
}

// Name implements Backend.
func (b *FluxBackend) Name() string {
	model := b.Model
	if model == "" {
		model = "schnell"
	}
	return "flux-" + model
}

// Txt2Img generates an image from a text prompt using FLUX.1.
func (b *FluxBackend) Txt2Img(ctx context.Context, req GenerateRequest) (*GenerateResult, error) {
	if req.Prompt == "" {
		return nil, fmt.Errorf("flux: prompt is required")
	}

	bin := b.BinaryPath
	if bin == "" {
		bin = "flux"
	}
	model := b.Model
	if model == "" {
		model = "schnell"
	}

	width := req.Width
	if width == 0 {
		width = 1024
	}
	height := req.Height
	if height == 0 {
		height = 1024
	}
	steps := req.Steps
	if steps == 0 {
		steps = 4
	}

	outPath := req.OutputPath
	if outPath == "" {
		outDir := b.OutputDir
		if outDir == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("flux: resolve home dir: %w", err)
			}
			outDir = filepath.Join(home, "generated-images")
		}
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return nil, fmt.Errorf("flux: create output dir: %w", err)
		}
		outPath = filepath.Join(outDir, fmt.Sprintf("%d.png", time.Now().UnixNano()))
	}

	args := []string{
		"--prompt", req.Prompt,
		"--width", fmt.Sprintf("%d", width),
		"--height", fmt.Sprintf("%d", height),
		"--steps", fmt.Sprintf("%d", steps),
		"--model", model,
		"--output", outPath,
	}
	if b.Quantize {
		args = append(args, "--quantize")
	}
	if req.Seed != 0 {
		args = append(args, "--seed", fmt.Sprintf("%d", req.Seed))
	}

	start := time.Now()
	cmd := exec.CommandContext(ctx, bin, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("flux: %w: %s", err, strings.TrimSpace(string(output)))
	}
	elapsed := time.Since(start)

	return &GenerateResult{
		OutputPath:  outPath,
		Width:       width,
		Height:      height,
		Steps:       steps,
		Model:       "black-forest-labs/FLUX.1-" + model,
		BackendName: b.Name(),
		Elapsed:     elapsed,
		Seed:        req.Seed,
	}, nil
}
