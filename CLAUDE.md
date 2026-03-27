@AGENTS.md

## Conventions
- `ImageGenerator` is the key interface; `FluxGenerator` is the production implementation
- `FluxGenerator` shells out to the `flux` CLI binary (flux.swift, installed separately)
- `ImageWorker` implements axon-task's `Worker` interface for async image generation
- `PromptMerger` combines baseline rules with scene prompts — do not bypass it
- Gallery storage uses filesystem with thumbnail variants

## Constraints
- Depends on axon-tool only — no dependency on axon server toolkit
- flux.swift CLI must be pre-installed (`just install-flux` in lamina root)
- Do not add LLM provider code — prompt merging is deterministic, not LLM-based
- Do not embed the flux binary or model weights — they are external dependencies

## Testing
- `go test ./...` — unit tests mock the ImageGenerator interface
- `go vet ./...` — must be clean
- Integration tests that call flux CLI are gated; they need the binary installed
