package lens

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
)

// ImageTaskParams holds parameters for an image generation task.
type ImageTaskParams struct {
	Prompt    string `json:"prompt"`
	AgentSlug string `json:"agent_slug"`
	UserID    string `json:"user_id"`
	ImageID   string `json:"image_id"`
}

// ImageWorker executes image generation tasks.
type ImageWorker struct {
	generator ImageGenerator
	images    *ImageStore
	gallery   GalleryStore
}

// NewImageWorker creates an ImageWorker with required dependencies.
func NewImageWorker(gen ImageGenerator, images *ImageStore, opts ...ImageWorkerOption) *ImageWorker {
	w := &ImageWorker{
		generator: gen,
		images:    images,
	}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

// ImageWorkerOption configures an ImageWorker during construction.
type ImageWorkerOption func(*ImageWorker)

// WithGallery sets the gallery store for recording generated images.
func WithGallery(g GalleryStore) ImageWorkerOption {
	return func(w *ImageWorker) {
		w.gallery = g
	}
}

// Execute generates an image and saves it. The params argument is
// JSON-encoded ImageTaskParams. This method signature is compatible
// with axon-task's Worker interface without importing it.
func (w *ImageWorker) Execute(ctx context.Context, params json.RawMessage) error {
	var p ImageTaskParams
	if err := json.Unmarshal(params, &p); err != nil {
		return fmt.Errorf("parse image task params: %w", err)
	}

	if p.Prompt == "" {
		return fmt.Errorf("empty prompt")
	}
	if p.ImageID == "" {
		return fmt.Errorf("empty image ID")
	}

	slog.Info("generating image", "image_id", p.ImageID, "prompt_len", len(p.Prompt))

	data, err := w.generator.GenerateImage(ctx, p.Prompt)
	if err != nil {
		return fmt.Errorf("generate image: %w", err)
	}

	if err := w.images.SaveWithID(p.ImageID, data); err != nil {
		return fmt.Errorf("save image: %w", err)
	}

	if w.gallery != nil {
		img := GalleryImage{
			ID:        p.ImageID,
			AgentSlug: p.AgentSlug,
			UserID:    p.UserID,
			Prompt:    p.Prompt,
		}
		if err := w.gallery.SaveGalleryImage(img); err != nil {
			slog.Warn("failed to save gallery record", "error", err, "image_id", p.ImageID)
		}
	}

	slog.Info("image generated", "image_id", p.ImageID)
	return nil
}
