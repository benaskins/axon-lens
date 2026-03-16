package lens

import (
	"context"
	"time"
)

// GalleryImage represents a generated image with metadata.
type GalleryImage struct {
	ID             string    `json:"id"`
	AgentSlug      string    `json:"agent_slug"`
	UserID         string    `json:"user_id"`
	ConversationID *string   `json:"conversation_id"`
	Prompt         string    `json:"prompt"`
	Model          string    `json:"model"`
	CreatedAt      time.Time `json:"created_at"`
}

// ImageTaskSubmission holds the parameters for an image generation task.
type ImageTaskSubmission struct {
	Prompt         string `json:"prompt"`
	AgentSlug      string `json:"agent_slug"`
	UserID         string `json:"user_id"`
	ConversationID string `json:"conversation_id,omitempty"`
	ImageID        string `json:"image_id"`
}

// TaskSubmitRequest is the request body for submitting a task.
type TaskSubmitRequest struct {
	Type   string `json:"type"`
	Params any    `json:"params"`
}

// TaskSubmission is the response from submitting a task.
type TaskSubmission struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
}

// NewImageTaskRequest creates a TaskSubmitRequest for image generation.
func NewImageTaskRequest(params *ImageTaskSubmission) *TaskSubmitRequest {
	return &TaskSubmitRequest{
		Type:   "image_generation",
		Params: params,
	}
}

// GalleryStore persists gallery images.
type GalleryStore interface {
	SaveGalleryImage(img GalleryImage) error
	GetGalleryImage(id string) (*GalleryImage, error)
	ListGalleryImagesByUser(userID string, agentSlug string) ([]GalleryImage, error)
}

// MessageStore provides access to recent conversation messages
// for prompt merging context.
type MessageStore interface {
	GetRecentMessages(conversationID string, limit int) ([]Message, error)
}

// Message is a minimal message type for prompt merging context.
type Message struct {
	Role    string
	Content string
}

// TaskSubmitter submits image generation tasks to an external runner.
type TaskSubmitter interface {
	SubmitTask(ctx context.Context, req *TaskSubmitRequest) (*TaskSubmission, error)
}

// CameraPrompt returns the system prompt section for the take_photo skill.
func CameraPrompt() string {
	return `## Camera
You have a camera and can take photos. Use it when there's a visual moment worth sharing — a new setting, something you made, or when the user asks to see something.`
}
