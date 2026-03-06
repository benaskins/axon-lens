package lens_test

import (
	"bytes"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	lens "github.com/benaskins/axon-lens"
)

func mustNewImageStore(t *testing.T, dir string) *lens.ImageStore {
	t.Helper()
	store, err := lens.NewImageStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func TestImageStore_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	store := mustNewImageStore(t, dir)

	data := []byte("fake png data")
	id, err := store.Save(data)
	if err != nil {
		t.Fatal(err)
	}
	if id == "" {
		t.Fatal("expected non-empty ID")
	}

	loaded, err := store.Load(id)
	if err != nil {
		t.Fatal(err)
	}
	if string(loaded) != string(data) {
		t.Errorf("loaded data = %q, want %q", loaded, data)
	}
}

func TestImageStore_SaveWithID(t *testing.T) {
	dir := t.TempDir()
	store := mustNewImageStore(t, dir)

	data := []byte("image bytes")
	err := store.SaveWithID("custom-id", data)
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := store.Load("custom-id")
	if err != nil {
		t.Fatal(err)
	}
	if string(loaded) != string(data) {
		t.Errorf("loaded = %q, want %q", loaded, data)
	}
}

func TestImageStore_LoadSize_Variant(t *testing.T) {
	dir := t.TempDir()
	store := mustNewImageStore(t, dir)

	// Write original and thumb variant
	os.WriteFile(filepath.Join(dir, "img1.png"), []byte("original"), 0644)
	os.WriteFile(filepath.Join(dir, "img1_thumb.png"), []byte("thumb"), 0644)

	data, err := store.LoadSize("img1", "thumb")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "thumb" {
		t.Errorf("got %q, want %q", data, "thumb")
	}
}

func TestImageStore_LoadSize_FallsBackToOriginal(t *testing.T) {
	dir := t.TempDir()
	store := mustNewImageStore(t, dir)

	os.WriteFile(filepath.Join(dir, "img1.png"), []byte("original"), 0644)

	// Request a size variant that doesn't exist
	data, err := store.LoadSize("img1", "medium")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "original" {
		t.Errorf("got %q, want %q", data, "original")
	}
}

func TestImageStore_LoadSize_RejectsPathTraversal(t *testing.T) {
	dir := t.TempDir()
	store := mustNewImageStore(t, dir)

	for _, id := range []string{"../etc/passwd", "foo/bar", "a\\b", "a..b/c"} {
		_, err := store.LoadSize(id, "")
		if err == nil {
			t.Errorf("expected error for ID %q", id)
		}
	}
}

func TestImageStore_Load_NotFound(t *testing.T) {
	dir := t.TempDir()
	store := mustNewImageStore(t, dir)

	_, err := store.Load("nonexistent")
	if err == nil {
		t.Error("expected error for missing image")
	}
}

func TestImageHandler_ServesImage(t *testing.T) {
	dir := t.TempDir()
	store := mustNewImageStore(t, dir)
	store.SaveWithID("test-img", []byte("png data"))

	handler := lens.ImageHandler(store)

	// Use a mux to set up PathValue
	mux := http.NewServeMux()
	mux.Handle("GET /api/images/{id}", handler)

	req := httptest.NewRequest("GET", "/api/images/test-img", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Header().Get("Content-Type") != "image/png" {
		t.Errorf("Content-Type = %q, want image/png", w.Header().Get("Content-Type"))
	}
	if w.Body.String() != "png data" {
		t.Errorf("body = %q, want %q", w.Body.String(), "png data")
	}
}

func TestImageHandler_NotFound(t *testing.T) {
	dir := t.TempDir()
	store := mustNewImageStore(t, dir)

	handler := lens.ImageHandler(store)
	mux := http.NewServeMux()
	mux.Handle("GET /api/images/{id}", handler)

	req := httptest.NewRequest("GET", "/api/images/missing", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestImageHandler_MethodNotAllowed(t *testing.T) {
	dir := t.TempDir()
	store := mustNewImageStore(t, dir)

	handler := lens.ImageHandler(store)

	req := httptest.NewRequest("POST", "/api/images/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func createTestPNG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestImageStore_SaveWithID_GeneratesThumbnails(t *testing.T) {
	dir := t.TempDir()
	store := mustNewImageStore(t, dir)

	data := createTestPNG(t, 2048, 1024)
	err := store.SaveWithID("img1", data)
	if err != nil {
		t.Fatal(err)
	}

	// Should have generated three thumbnail variants
	variants := []struct {
		suffix  string
		maxSide int
	}{
		{"_thumb", 256},
		{"_medium", 512},
		{"_lg", 1024},
	}

	for _, v := range variants {
		path := filepath.Join(dir, "img1"+v.suffix+".png")
		thumbData, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("expected %s variant to exist: %v", v.suffix, err)
			continue
		}

		img, err := png.Decode(bytes.NewReader(thumbData))
		if err != nil {
			t.Errorf("failed to decode %s variant: %v", v.suffix, err)
			continue
		}

		bounds := img.Bounds()
		longestSide := bounds.Dx()
		if bounds.Dy() > longestSide {
			longestSide = bounds.Dy()
		}
		if longestSide != v.maxSide {
			t.Errorf("%s: longest side = %d, want %d", v.suffix, longestSide, v.maxSide)
		}
	}
}

func TestImageStore_SaveWithID_SkipsThumbnailsForSmallImages(t *testing.T) {
	dir := t.TempDir()
	store := mustNewImageStore(t, dir)

	// Image smaller than all variant sizes — variants should be copies
	data := createTestPNG(t, 100, 80)
	err := store.SaveWithID("small", data)
	if err != nil {
		t.Fatal(err)
	}

	// All variants should exist and match the original size
	for _, suffix := range []string{"_thumb", "_medium", "_lg"} {
		path := filepath.Join(dir, "small"+suffix+".png")
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected %s variant to exist", suffix)
		}
	}
}

func TestImageStore_BackfillThumbnails(t *testing.T) {
	dir := t.TempDir()
	store := mustNewImageStore(t, dir)

	// Write a PNG directly without going through SaveWithID
	data := createTestPNG(t, 2048, 1024)
	os.WriteFile(filepath.Join(dir, "backfill-test.png"), data, 0644)

	count, err := store.BackfillThumbnails()
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("backfilled %d, want 1", count)
	}

	// Variants should now exist
	for _, suffix := range []string{"_thumb", "_medium", "_lg"} {
		path := filepath.Join(dir, "backfill-test"+suffix+".png")
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected %s variant after backfill", suffix)
		}
	}

	// Running again should find nothing to do
	count, err = store.BackfillThumbnails()
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("second backfill processed %d, want 0", count)
	}
}

func TestImageStore_SaveWithID_RejectsPathTraversal(t *testing.T) {
	dir := t.TempDir()
	store := mustNewImageStore(t, dir)

	for _, id := range []string{"../etc/passwd", "foo/bar", "a\\b", "a..b/c"} {
		err := store.SaveWithID(id, []byte("data"))
		if err == nil {
			t.Errorf("expected error for ID %q", id)
		}
	}
}
