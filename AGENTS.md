# axon-lens

Image generation pipeline — prompt merging, FLUX.1 via MLX, gallery storage.

## Build & Test

```bash
go test ./...
go vet ./...
```

## Key Files

- `generate.go` — FluxGenerator and ImageGenerator interface
- `image_store.go` — ImageStore with filesystem storage and thumbnail variants
- `handler_gallery.go` — GalleryListHandler and ImageHandler HTTP handlers
- `photo.go` — TakePhotoTool definition and CameraPrompt
- `prompt_merge.go` — PromptMerger for combining baseline rules with scene prompts
- `tools.go` — Config, TaskSubmitter, and task submission types
- `worker.go` — ImageWorker for axon-task integration
- `doc.go` — package documentation
