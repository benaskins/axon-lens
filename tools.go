package lens

import (
	"context"
	"encoding/base64"
	"log/slog"

	"github.com/google/uuid"

	tool "github.com/benaskins/axon-tool"
)

// Config holds dependencies for building photo tools.
type Config struct {
	TaskSubmitter TaskSubmitter
	PromptMerger  *PromptMerger  // nil = no prompt merging
	ImageStore    *ImageStore    // nil = no ref image loading
	GalleryStore  GalleryStore   // nil = no base image
	MessageStore  MessageStore   // nil = no recent message context
	OnTaskStarted func(taskID, taskType, description string)
	StartPoll     func(taskID string)
}

// TakePhotoTool returns a tool.ToolDef for the take_photo skill.
func TakePhotoTool(cfg *Config) tool.ToolDef {
	return tool.ToolDef{
		Name:        "take_photo",
		Description: "Take a photo or selfie. Only use when the user asks, you're in a new setting, or a significant visual moment happens. Most conversation turns should NOT include a photo.",
		Parameters: tool.ParameterSchema{
			Type:     "object",
			Required: []string{"prompt"},
			Properties: map[string]tool.PropertySchema{
				"prompt": {Type: "string", Description: "A detailed description of the image to generate. Include subject, setting, lighting, mood, and composition details."},
			},
		},
		Execute: func(ctx *tool.ToolContext, args map[string]any) tool.ToolResult {
			promptStr, _ := args["prompt"].(string)
			if promptStr == "" || cfg.TaskSubmitter == nil {
				return tool.ToolResult{Content: "Error: image generation not available"}
			}
			return submitImageTask(cfg, ctx, promptStr)
		},
	}
}

func submitImageTask(cfg *Config, ctx *tool.ToolContext, promptStr string) tool.ToolResult {
	imageID := uuid.New().String()

	finalPrompt := promptStr
	if cfg.PromptMerger != nil {
		var recentMessages []Message
		if ctx.ConversationID != "" && cfg.MessageStore != nil {
			if msgs, err := cfg.MessageStore.GetRecentMessages(ctx.ConversationID, 5); err == nil {
				recentMessages = msgs
			}
		}

		merged, err := cfg.PromptMerger.MergePrompt(ctx.SystemPrompt, recentMessages, promptStr)
		if err == nil {
			finalPrompt = merged
		} else {
			slog.Warn("prompt merge failed, using raw scene prompt", "error", err)
		}
	}

	var refImageB64 string
	if ctx.AgentSlug != "" && cfg.GalleryStore != nil && cfg.ImageStore != nil {
		if baseImg, err := cfg.GalleryStore.GetBaseImageByUser(ctx.UserID, ctx.AgentSlug); err == nil && baseImg != nil {
			if imgData, err := cfg.ImageStore.Load(baseImg.ID); err == nil {
				refImageB64 = base64.StdEncoding.EncodeToString(imgData)
			} else {
				slog.Warn("failed to load base image", "error", err, "base_image_id", baseImg.ID)
			}
		}
	}

	submission := &ImageTaskSubmission{
		Prompt:         finalPrompt,
		ReferenceImage: refImageB64,
		AgentSlug:      ctx.AgentSlug,
		UserID:         ctx.UserID,
		ConversationID: ctx.ConversationID,
		ImageID:        imageID,
	}

	_, err := cfg.TaskSubmitter.SubmitTask(context.Background(), NewImageTaskRequest(submission))
	if err != nil {
		slog.Error("failed to submit image task", "error", err, "image_id", imageID)
		return tool.ToolResult{Content: "Error: failed to submit image generation task"}
	}

	if cfg.OnTaskStarted != nil {
		cfg.OnTaskStarted(imageID, "image_generation", "Generating image...")
	}

	if cfg.StartPoll != nil {
		cfg.StartPoll(imageID)
	}

	return tool.ToolResult{Content: "Image generation started, it will appear shortly."}
}
