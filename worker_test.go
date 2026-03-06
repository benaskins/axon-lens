package lens_test

import (
	"context"
	"encoding/json"
	"testing"

	lens "github.com/benaskins/axon-lens"
)

type stubImageGen struct {
	data []byte
	err  error
}

func (f *stubImageGen) GenerateImage(_ context.Context, _ string) ([]byte, error) {
	return f.data, f.err
}

type fakeGalleryStore struct {
	saved []lens.GalleryImage
}

func (f *fakeGalleryStore) SaveGalleryImage(img lens.GalleryImage) error {
	f.saved = append(f.saved, img)
	return nil
}

func (f *fakeGalleryStore) GetGalleryImage(id string) (*lens.GalleryImage, error) {
	return nil, nil
}

func (f *fakeGalleryStore) ListGalleryImagesByUser(userID, agentSlug string) ([]lens.GalleryImage, error) {
	return nil, nil
}

func (f *fakeGalleryStore) GetBaseImageByUser(userID, agentSlug string) (*lens.GalleryImage, error) {
	return nil, nil
}

func (f *fakeGalleryStore) SetBaseImage(userID, agentSlug, imageID string) error {
	return nil
}

func TestImageWorker_Execute(t *testing.T) {
	dir := t.TempDir()
	imgStore := mustNewImageStore(t, dir)
	pngData := createTestPNG(t, 512, 512)

	gallery := &fakeGalleryStore{}
	gen := &stubImageGen{data: pngData}

	worker := &lens.ImageWorker{
		Generator: gen,
		Images:    imgStore,
		Gallery:   gallery,
	}

	params := lens.ImageTaskParams{
		Prompt:    "a test image",
		AgentSlug: "test-agent",
		UserID:    "user1",
		ImageID:   "img-001",
	}
	raw, _ := json.Marshal(params)

	err := worker.Execute(context.Background(), raw)
	if err != nil {
		t.Fatal(err)
	}

	// Image should be saved
	loaded, err := imgStore.Load("img-001")
	if err != nil {
		t.Fatal("image not saved:", err)
	}
	if len(loaded) == 0 {
		t.Error("saved image is empty")
	}

	// Gallery record should be saved
	if len(gallery.saved) != 1 {
		t.Fatalf("expected 1 gallery record, got %d", len(gallery.saved))
	}
	if gallery.saved[0].ID != "img-001" {
		t.Errorf("gallery image ID = %q, want %q", gallery.saved[0].ID, "img-001")
	}
	if gallery.saved[0].AgentSlug != "test-agent" {
		t.Errorf("gallery agent = %q, want %q", gallery.saved[0].AgentSlug, "test-agent")
	}
}

func TestImageWorker_Execute_EmptyPrompt(t *testing.T) {
	worker := &lens.ImageWorker{}

	params := lens.ImageTaskParams{ImageID: "img-001"}
	raw, _ := json.Marshal(params)

	err := worker.Execute(context.Background(), raw)
	if err == nil {
		t.Error("expected error for empty prompt")
	}
}

func TestImageWorker_Execute_EmptyImageID(t *testing.T) {
	worker := &lens.ImageWorker{}

	params := lens.ImageTaskParams{Prompt: "test"}
	raw, _ := json.Marshal(params)

	err := worker.Execute(context.Background(), raw)
	if err == nil {
		t.Error("expected error for empty image ID")
	}
}

func TestImageWorker_Execute_NilGallery(t *testing.T) {
	dir := t.TempDir()
	imgStore := mustNewImageStore(t, dir)
	pngData := createTestPNG(t, 64, 64)

	worker := &lens.ImageWorker{
		Generator: &stubImageGen{data: pngData},
		Images:    imgStore,
		Gallery:   nil,
	}

	params := lens.ImageTaskParams{
		Prompt:  "test",
		ImageID: "img-002",
	}
	raw, _ := json.Marshal(params)

	err := worker.Execute(context.Background(), raw)
	if err != nil {
		t.Fatal("should succeed without gallery store:", err)
	}
}
