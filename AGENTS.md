# axon-lens

Image generation pipeline — prompt merging, FLUX.1 via MLX, gallery storage.

## Build & Test

```bash
go test ./...
go vet ./...
```

## Key Files

- `generate.go` — image generation orchestration
- `handler_gallery.go` — gallery HTTP handlers
- `image_store.go` — image storage backend
