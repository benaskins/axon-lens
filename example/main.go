//go:build ignore

// Example: generate an image with SDXL-Turbo or FLUX.1 via the Backend interface.
//
// Usage:
//
//	go run example/main.go               # SDXL-Turbo, 512x512
//	go run example/main.go flux          # FLUX.1 Schnell, 1024x1024
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	lens "github.com/benaskins/axon-lens"
)

func main() {
	var backend lens.Backend
	req := lens.GenerateRequest{
		Prompt: "a red fox sitting in a snowy forest at dusk",
		Width:  512,
		Height: 512,
		Steps:  1,
	}

	if len(os.Args) > 1 && os.Args[1] == "flux" {
		backend = lens.NewFluxBackend()
		req.Width = 1024
		req.Height = 1024
		req.Steps = 4
	} else {
		backend = lens.NewSDXLBackend()
	}

	ctx := context.Background()
	result, err := backend.Txt2Img(ctx, req)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("backend:  %s\n", result.BackendName)
	fmt.Printf("model:    %s\n", result.Model)
	fmt.Printf("size:     %dx%d\n", result.Width, result.Height)
	fmt.Printf("elapsed:  %s\n", result.Elapsed)
	fmt.Printf("output:   %s\n", result.OutputPath)
}
