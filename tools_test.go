package lens_test

import (
	"context"
	"testing"

	lens "github.com/benaskins/axon-lens"
	tool "github.com/benaskins/axon-tool"
)

type mockTaskSubmitter struct {
	lastReq *lens.TaskSubmitRequest
	lastCtx context.Context
}

func (m *mockTaskSubmitter) SubmitTask(ctx context.Context, req *lens.TaskSubmitRequest) (*lens.TaskSubmission, error) {
	m.lastReq = req
	m.lastCtx = ctx
	return &lens.TaskSubmission{TaskID: "task-1", Status: "queued"}, nil
}

func TestTakePhotoTool_Schema(t *testing.T) {
	cfg := &lens.Config{TaskSubmitter: &mockTaskSubmitter{}}
	td := lens.TakePhotoTool(cfg)

	if td.Name != "take_photo" {
		t.Errorf("Name = %q, want %q", td.Name, "take_photo")
	}
	if _, ok := td.Parameters.Properties["prompt"]; !ok {
		t.Error("expected prompt parameter")
	}
}

func TestTakePhotoTool_SubmitsTask(t *testing.T) {
	submitter := &mockTaskSubmitter{}
	var startedTaskID, startedType string
	cfg := &lens.Config{
		TaskSubmitter: submitter,
		OnTaskStarted: func(taskID, taskType, desc string) {
			startedTaskID = taskID
			startedType = taskType
		},
	}

	td := lens.TakePhotoTool(cfg)
	ctx := &tool.ToolContext{
		Ctx:       context.Background(),
		UserID:    "user-1",
		AgentSlug: "bot",
	}

	result := td.Execute(ctx, map[string]any{"prompt": "a sunset"})
	if result.Content == "" {
		t.Error("expected non-empty result")
	}

	if submitter.lastReq == nil {
		t.Fatal("expected task to be submitted")
	}
	if submitter.lastReq.Type != "image_generation" {
		t.Errorf("Type = %q, want %q", submitter.lastReq.Type, "image_generation")
	}

	if startedTaskID == "" {
		t.Error("expected OnTaskStarted to be called")
	}
	if startedType != "image_generation" {
		t.Errorf("startedType = %q, want %q", startedType, "image_generation")
	}
}

func TestTakePhotoTool_EmptyPrompt(t *testing.T) {
	cfg := &lens.Config{TaskSubmitter: &mockTaskSubmitter{}}
	td := lens.TakePhotoTool(cfg)

	result := td.Execute(&tool.ToolContext{Ctx: context.Background()}, map[string]any{"prompt": ""})
	if result.Content != "Error: image generation not available" {
		t.Errorf("result = %q", result.Content)
	}
}

func TestTakePhotoTool_NoSubmitter(t *testing.T) {
	cfg := &lens.Config{} // no TaskSubmitter
	td := lens.TakePhotoTool(cfg)

	result := td.Execute(&tool.ToolContext{Ctx: context.Background()}, map[string]any{"prompt": "test"})
	if result.Content != "Error: image generation not available" {
		t.Errorf("result = %q", result.Content)
	}
}

func TestTakePhotoTool_WithPromptMerger(t *testing.T) {
	submitter := &mockTaskSubmitter{}
	gen := fakeGenerator("merged prompt")
	merger := lens.NewPromptMerger(gen, &lens.ImageGenConfig{
		MergeInstruction: "{scene}",
	})

	cfg := &lens.Config{
		TaskSubmitter: submitter,
		PromptMerger:  merger,
	}

	td := lens.TakePhotoTool(cfg)
	result := td.Execute(&tool.ToolContext{Ctx: context.Background()}, map[string]any{"prompt": "raw scene"})

	if result.Content == "" {
		t.Error("expected non-empty result")
	}
	if submitter.lastReq == nil {
		t.Fatal("expected task submission")
	}
	params := submitter.lastReq.Params.(*lens.ImageTaskSubmission)
	if params.Prompt != "merged prompt" {
		t.Errorf("Prompt = %q, want %q", params.Prompt, "merged prompt")
	}
}

func TestTakePhotoTool_PropagatesCallerContext(t *testing.T) {
	type ctxKey string
	submitter := &mockTaskSubmitter{}
	cfg := &lens.Config{TaskSubmitter: submitter}

	td := lens.TakePhotoTool(cfg)
	callerCtx := context.WithValue(context.Background(), ctxKey("trace"), "abc123")
	ctx := &tool.ToolContext{
		Ctx:    callerCtx,
		UserID: "user-1",
	}

	td.Execute(ctx, map[string]any{"prompt": "test"})

	if submitter.lastCtx == nil {
		t.Fatal("expected context to be passed to SubmitTask")
	}
	if submitter.lastCtx.Value(ctxKey("trace")) != "abc123" {
		t.Error("expected caller context to be propagated, not context.Background()")
	}
}

func TestTakePhotoTool_StartPollCalled(t *testing.T) {
	var polledID string
	cfg := &lens.Config{
		TaskSubmitter: &mockTaskSubmitter{},
		StartPoll: func(taskID string) {
			polledID = taskID
		},
	}

	td := lens.TakePhotoTool(cfg)
	td.Execute(&tool.ToolContext{Ctx: context.Background()}, map[string]any{"prompt": "test"})

	if polledID == "" {
		t.Error("expected StartPoll to be called")
	}
}
