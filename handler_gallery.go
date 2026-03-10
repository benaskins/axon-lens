package lens

import (
	"context"
	"encoding/json"
	"net/http"
)

// UserIDFunc extracts the authenticated user ID from a request context.
type UserIDFunc func(ctx context.Context) string

// writeJSON writes v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// GalleryListHandler returns an http.Handler for GET /api/agents/{slug}/gallery.
func GalleryListHandler(store GalleryStore, userID UserIDFunc) http.Handler {
	return &galleryListHandler{store: store, userID: userID}
}

type galleryListHandler struct {
	store  GalleryStore
	userID UserIDFunc
}

func (h *galleryListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid := h.userID(r.Context())
	if uid == "" {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	slug := r.PathValue("slug")
	if slug == "" {
		writeError(w, http.StatusBadRequest, "slug required")
		return
	}

	images, err := h.store.ListGalleryImagesByUser(uid, slug)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list images")
		return
	}

	if images == nil {
		images = []GalleryImage{}
	}

	type imageResponse struct {
		ID             string  `json:"id"`
		URL            string  `json:"url"`
		ThumbnailURL   string  `json:"thumbnail_url"`
		Prompt         string  `json:"prompt"`
		Model          string  `json:"model"`
		ConversationID *string `json:"conversation_id"`
		CreatedAt      string  `json:"created_at"`
	}

	response := make([]imageResponse, len(images))
	for i, img := range images {
		response[i] = imageResponse{
			ID:             img.ID,
			URL:            "/api/images/" + img.ID,
			ThumbnailURL:   "/api/images/" + img.ID + "?size=thumb",
			Prompt:         img.Prompt,
			Model:          img.Model,
			ConversationID: img.ConversationID,
			CreatedAt:      img.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{"images": response})
}
