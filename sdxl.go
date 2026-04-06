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

// SDXLBackend generates images by calling the generate-image shell script,
// which uses SDXL-Turbo via PyTorch MPS (Apple Silicon GPU).
//
// The generate-image script must be installed and executable. The default
// lookup uses PATH; override ScriptPath for an explicit location.
type SDXLBackend struct {
	// ScriptPath is the path to the generate-image script.
	// Defaults to "generate-image" (resolved via PATH).
	ScriptPath string

	// OutputDir is the default directory for generated images.
	// Defaults to ~/generated-images (matching the script's default).
	OutputDir string
}

// NewSDXLBackend returns an SDXLBackend that resolves generate-image via PATH.
func NewSDXLBackend() *SDXLBackend {
	return &SDXLBackend{ScriptPath: "generate-image"}
}

// Name implements Backend.
func (b *SDXLBackend) Name() string { return "sdxl-turbo" }

// Txt2Img generates an image from a text prompt using SDXL-Turbo.
// The script prints the output path to stdout; GenerateResult.OutputPath
// is set to that value.
func (b *SDXLBackend) Txt2Img(ctx context.Context, req GenerateRequest) (*GenerateResult, error) {
	if req.Prompt == "" {
		return nil, fmt.Errorf("sdxl: prompt is required")
	}

	script := b.ScriptPath
	if script == "" {
		script = "generate-image"
	}

	width := req.Width
	if width == 0 {
		width = 512
	}
	height := req.Height
	if height == 0 {
		height = 512
	}
	steps := req.Steps
	if steps == 0 {
		steps = 1
	}

	outPath := req.OutputPath
	if outPath == "" {
		outDir := b.OutputDir
		if outDir == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("sdxl: resolve home dir: %w", err)
			}
			outDir = filepath.Join(home, "generated-images")
		}
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return nil, fmt.Errorf("sdxl: create output dir: %w", err)
		}
		outPath = filepath.Join(outDir, fmt.Sprintf("%d.png", time.Now().UnixNano()))
	}

	args := []string{
		"--output", outPath,
		"--steps", fmt.Sprintf("%d", steps),
		"--width", fmt.Sprintf("%d", width),
		"--height", fmt.Sprintf("%d", height),
		"--backend", "sdxl",
	}
	if req.Seed != 0 {
		args = append(args, "--seed", fmt.Sprintf("%d", req.Seed))
	}
	args = append(args, req.Prompt)

	start := time.Now()
	cmd := exec.CommandContext(ctx, script, args...)
	out, err := cmd.Output()
	elapsed := time.Since(start)
	if err != nil {
		stderr := ""
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = string(exitErr.Stderr)
		}
		return nil, fmt.Errorf("sdxl: generate-image failed: %w\n%s", err, strings.TrimSpace(stderr))
	}

	// The script prints the output path to stdout.
	resultPath := strings.TrimSpace(string(out))
	if resultPath == "" {
		resultPath = outPath
	}

	return &GenerateResult{
		OutputPath:  resultPath,
		Width:       width,
		Height:      height,
		Steps:       steps,
		Model:       "stabilityai/sdxl-turbo",
		BackendName: b.Name(),
		Elapsed:     elapsed,
		Seed:        req.Seed,
	}, nil
}
