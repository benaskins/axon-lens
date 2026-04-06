package lens

import (
	"context"
	"time"
)

// GenerateRequest holds parameters for a single image generation operation.
// Zero values are interpreted as "use backend default".
type GenerateRequest struct {
	// Prompt is the text description of the image to generate. Required.
	Prompt string

	// Width and Height in pixels. Defaults depend on the backend.
	Width  int
	Height int

	// Steps is the number of inference steps. Defaults depend on the backend.
	Steps int

	// Seed for reproducibility. 0 means no seed (random).
	Seed int64

	// OutputPath is the file path to write the generated image to.
	// If empty the backend chooses a path under its default output directory.
	OutputPath string
}

// GenerateResult holds the outcome of a completed generation operation.
type GenerateResult struct {
	// OutputPath is the absolute path to the generated image file.
	OutputPath string

	// Width and Height of the generated image in pixels.
	Width  int
	Height int

	// Steps used during generation.
	Steps int

	// Model is the model identifier used (e.g. "stabilityai/sdxl-turbo").
	Model string

	// BackendName is the Backend.Name() that produced this result.
	BackendName string

	// Elapsed is the wall-clock duration of the generation call.
	Elapsed time.Duration

	// Seed used (0 if none was specified).
	Seed int64
}

// Backend performs image generation operations against a specific backend.
// Implementations wrap external tools (generate-image script, flux.swift CLI, etc.).
// Backends are designed to be composable: results can be passed between operations
// (e.g. txt2img → img2img) using the OutputPath field.
type Backend interface {
	// Name returns a stable identifier for this backend (e.g. "sdxl-turbo", "flux-schnell").
	Name() string

	// Txt2Img generates an image from a text prompt.
	Txt2Img(ctx context.Context, req GenerateRequest) (*GenerateResult, error)
}
