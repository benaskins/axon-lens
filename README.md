# axon-lens

Image generation tools for LLM-powered agents. Part of [lamina](https://github.com/benaskins/lamina) — each axon package can be used independently.

Handles prompt merging, image storage, gallery management, and task submission for image generation pipelines.

## Install

```
go get github.com/benaskins/axon-lens@latest
```

Requires Go 1.24+.

## Usage

```go
cfg := &lens.Config{
    PromptMerger: promptMerger,
    ImageStore:   imageStore,
    GalleryStore: galleryStore,
    MessageStore: messageStore,
}

tools := map[string]tool.ToolDef{
    "take_photo": lens.TakePhotoTool(cfg),
}
```

### Key types

- `Config` — wires together prompt merging, storage, and task submission
- `TakePhotoTool()` — tool constructor for LLM agents
- `PromptMerger` — merges user prompts with style/character configuration
- `ImageStore` — local image file storage with thumbnails
- `GalleryStore` — persistence interface for gallery images
- `ImageHandler()`, `GalleryListHandler()` — HTTP handlers for serving images

## License

MIT — see [LICENSE](LICENSE).
