# axon-lens

> Primitives · Part of the [lamina](https://github.com/benaskins/lamina-mono) workspace

Image generation pipeline for Apple Silicon. Provides an `ImageGenerator` interface with a FLUX.1 implementation via [flux.swift](https://github.com/filipstrand/mluX), local image storage with automatic thumbnail variants, gallery management, LLM-based prompt merging, and a ready-made tool definition for LLM agents.

## Getting started

```bash
go get github.com/benaskins/axon-lens@latest
```

Requires Go 1.24+.

```go
gen := &lens.FluxGenerator{
    BinaryPath: "flux",
    Model:      "schnell",
    Width:      1024,
    Height:     1024,
    Steps:      4,
}

data, err := gen.GenerateImage(ctx, "a red fox in a snowy forest")
if err != nil {
    log.Fatal(err)
}

store, _ := lens.NewImageStore("/tmp/images")
id, _ := store.Save(data)
fmt.Println("saved:", id)
```

## Key types

- **`ImageGenerator`** — interface for image generation backends (`GenerateImage(ctx, prompt) ([]byte, error)`)
- **`FluxGenerator`** — FLUX.1 implementation that shells out to the flux.swift CLI
- **`ImageStore`** — filesystem storage with automatic thumbnail generation (256px, 512px, 1024px variants)
- **`ImageWorker`** — task worker compatible with axon-task (generate, save, record in gallery)
- **`PromptMerger`** — merges baseline rules, agent context, and scene descriptions via an LLM
- **`TakePhotoTool(cfg)`** — returns an axon-tool `ToolDef` for LLM agents (requires a `*Config`)
- **`GalleryStore`** — interface for persisting gallery image metadata
- **`GalleryListHandler()`** / **`ImageHandler()`** — HTTP handlers for serving images and gallery listings

## License

MIT
