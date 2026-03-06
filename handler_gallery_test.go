package lens_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	lens "github.com/benaskins/axon-lens"
)

func testUserID(ctx context.Context) string {
	return "user-1"
}

// memGalleryStore is a minimal in-memory GalleryStore for testing.
type memGalleryStore struct {
	images map[string]lens.GalleryImage
}

func newMemGalleryStore() *memGalleryStore {
	return &memGalleryStore{
		images: make(map[string]lens.GalleryImage),
	}
}

func (s *memGalleryStore) SaveGalleryImage(img lens.GalleryImage) error {
	s.images[img.ID] = img
	return nil
}

func (s *memGalleryStore) GetGalleryImage(id string) (*lens.GalleryImage, error) {
	img, ok := s.images[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return &img, nil
}

func (s *memGalleryStore) ListGalleryImagesByUser(userID, slug string) ([]lens.GalleryImage, error) {
	var result []lens.GalleryImage
	for _, img := range s.images {
		if img.UserID == userID && img.AgentSlug == slug {
			result = append(result, img)
		}
	}
	return result, nil
}

func TestGalleryListHandler_ReturnsImages(t *testing.T) {
	store := newMemGalleryStore()
	store.SaveGalleryImage(lens.GalleryImage{
		ID:        "img-1",
		AgentSlug: "bot",
		UserID:    "user-1",
		Prompt:    "sunset",
		Model:     "flux-schnell",
		CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	})

	handler := lens.GalleryListHandler(store, testUserID)
	mux := http.NewServeMux()
	mux.Handle("GET /api/agents/{slug}/gallery", handler)

	req := httptest.NewRequest("GET", "/api/agents/bot/gallery", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !contains(body, "img-1") {
		t.Errorf("body missing image ID: %s", body)
	}
	if !contains(body, "/api/images/img-1") {
		t.Errorf("body missing image URL: %s", body)
	}
}

func TestGalleryListHandler_EmptyGallery(t *testing.T) {
	store := newMemGalleryStore()
	handler := lens.GalleryListHandler(store, testUserID)
	mux := http.NewServeMux()
	mux.Handle("GET /api/agents/{slug}/gallery", handler)

	req := httptest.NewRequest("GET", "/api/agents/bot/gallery", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !contains(body, `"images":[]`) {
		t.Errorf("expected empty images array, got: %s", body)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
