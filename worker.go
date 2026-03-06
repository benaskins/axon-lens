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
	Generator    ImageGenerator
	Images       *ImageStore
	Gallery      GalleryStore
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

	data, err := w.Generator.GenerateImage(ctx, p.Prompt)
	if err != nil {
		return fmt.Errorf("generate image: %w", err)
	}

	if err := w.Images.SaveWithID(p.ImageID, data); err != nil {
		return fmt.Errorf("save image: %w", err)
	}

	if w.Gallery != nil {
		img := GalleryImage{
			ID:        p.ImageID,
			AgentSlug: p.AgentSlug,
			UserID:    p.UserID,
			Prompt:    p.Prompt,
		}
		if err := w.Gallery.SaveGalleryImage(img); err != nil {
			slog.Warn("failed to save gallery record", "error", err, "image_id", p.ImageID)
		}
	}

	slog.Info("image generated", "image_id", p.ImageID)
	return nil
}
