//go:build ignore

// Example: generate an image with FLUX.1 and save it to local storage.
package main

import (
	"context"
	"fmt"
	"log"

	lens "github.com/benaskins/axon-lens"
)

func main() {
	// Configure the FLUX.1 generator (requires flux.swift CLI installed).
	gen := &lens.FluxGenerator{
		BinaryPath: "flux",
		Model:      "schnell",
		Width:      1024,
		Height:     1024,
		Steps:      4,
		Quantize:   true,
	}

	// Generate an image from a text prompt.
	ctx := context.Background()
	data, err := gen.GenerateImage(ctx, "a red fox sitting in a snowy forest at dusk")
	if err != nil {
		log.Fatal(err)
	}

	// Save to local storage (creates thumbnails automatically).
	store, err := lens.NewImageStore("/tmp/axon-lens-images")
	if err != nil {
		log.Fatal(err)
	}

	id, err := store.Save(data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("saved image: %s\n", id)
}
