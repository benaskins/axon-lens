package lens_test

import (
	"context"
	"testing"
	"time"

	lens "github.com/benaskins/axon-lens"
)

func TestGalleryImageFields(t *testing.T) {
	img := lens.GalleryImage{
		ID:        "img-1",
		AgentSlug: "helper",
		UserID:    "user-1",
		Prompt:    "a sunset",
		Model:     "flux-schnell",
		CreatedAt: time.Now(),
	}

	if img.ID != "img-1" {
		t.Errorf("ID = %q, want %q", img.ID, "img-1")
	}
	if img.AgentSlug != "helper" {
		t.Errorf("AgentSlug = %q, want %q", img.AgentSlug, "helper")
	}
}

func TestImageTaskSubmissionFields(t *testing.T) {
	sub := lens.ImageTaskSubmission{
		Prompt:    "a sunset over mountains",
		AgentSlug: "helper",
		UserID:    "user-1",
		ImageID:   "img-1",
	}

	if sub.Prompt != "a sunset over mountains" {
		t.Errorf("Prompt = %q, want %q", sub.Prompt, "a sunset over mountains")
	}
}

// Verify interfaces are satisfied by simple implementations.

type stubGalleryStore struct{}

func (s *stubGalleryStore) SaveGalleryImage(img lens.GalleryImage) error { return nil }
func (s *stubGalleryStore) GetGalleryImage(id string) (*lens.GalleryImage, error) {
	return nil, nil
}
func (s *stubGalleryStore) ListGalleryImagesByUser(userID, slug string) ([]lens.GalleryImage, error) {
	return nil, nil
}

type stubTaskSubmitter struct{}

func (s *stubTaskSubmitter) SubmitTask(ctx context.Context, req *lens.TaskSubmitRequest) (*lens.TaskSubmission, error) {
	return &lens.TaskSubmission{TaskID: "t-1", Status: "queued"}, nil
}

func TestGalleryStoreInterface(t *testing.T) {
	var store lens.GalleryStore = &stubGalleryStore{}
	if err := store.SaveGalleryImage(lens.GalleryImage{}); err != nil {
		t.Fatal(err)
	}
}

func TestTaskSubmitterInterface(t *testing.T) {
	var sub lens.TaskSubmitter = &stubTaskSubmitter{}
	result, err := sub.SubmitTask(context.Background(), &lens.TaskSubmitRequest{Type: "image_generation"})
	if err != nil {
		t.Fatal(err)
	}
	if result.TaskID != "t-1" {
		t.Errorf("TaskID = %q, want %q", result.TaskID, "t-1")
	}
}
