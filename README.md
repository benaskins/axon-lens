# axon-lens

Image generation pipeline for LLM-powered agents. Part of [lamina](https://github.com/benaskins/lamina) — each axon package can be used independently.

Generates images via FLUX.1 on Apple Silicon (MLX), with prompt merging, thumbnail generation, and gallery management.

## Install

```
go get github.com/benaskins/axon-lens@latest
```

Requires Go 1.24+.

## Usage

### As an LLM tool

```go
cfg := &lens.Config{
    TaskSubmitter: submitter,
    PromptMerger:  promptMerger,
}

tools := map[string]tool.ToolDef{
    "take_photo": lens.TakePhotoTool(cfg),
}
```

### As a task worker

```go
fluxGen := &lens.FluxGenerator{
    BinaryPath: "flux",
    Model:      "schnell",
    Width:      1024,
    Height:     1024,
    Steps:      4,
    Quantize:   true,
}

worker := &lens.ImageWorker{
    Generator: fluxGen,
    Images:    imgStore,
    Gallery:   galleryStore,
}

executor.RegisterWorker("image_generation", worker)
```

### Key types

- `ImageGenerator` — interface for image generation backends
- `FluxGenerator` — FLUX.1 implementation via flux.swift CLI
- `ImageWorker` — task worker (generate → save → gallery record)
- `ImageStore` — local image file storage with thumbnails
- `PromptMerger` — LLM-based prompt merging with baseline rules
- `TakePhotoTool()` — tool constructor for LLM agents
- `GalleryListHandler()` — HTTP handler for gallery listing

## License

MIT — see [LICENSE](LICENSE).
