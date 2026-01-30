package main

import (
	"context"
	"strings"
	"testing"
	"time"

	mcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestToolTimeoutMiddleware(t *testing.T) {
	t.Run("applies deadline to tools/call", func(t *testing.T) {
		timeout := 200 * time.Millisecond
		middleware := toolTimeoutMiddleware(timeout)

		var capturedCtx context.Context
		handler := middleware(func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			capturedCtx = ctx
			return nil, nil
		})

		_, _ = handler(context.Background(), "tools/call", nil)

		deadline, ok := capturedCtx.Deadline()
		if !ok {
			t.Fatal("expected context to have a deadline for tools/call")
		}
		if time.Until(deadline) > timeout {
			t.Fatalf("deadline too far in the future: %v", time.Until(deadline))
		}
	})

	t.Run("no deadline for other methods", func(t *testing.T) {
		middleware := toolTimeoutMiddleware(200 * time.Millisecond)

		var capturedCtx context.Context
		handler := middleware(func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			capturedCtx = ctx
			return nil, nil
		})

		_, _ = handler(context.Background(), "initialize", nil)

		if _, ok := capturedCtx.Deadline(); ok {
			t.Fatal("expected no deadline for non-tools/call method")
		}
	})

	t.Run("timeout returns LLM-friendly error result", func(t *testing.T) {
		middleware := toolTimeoutMiddleware(10 * time.Millisecond)

		handler := middleware(func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			time.Sleep(50 * time.Millisecond)
			return nil, ctx.Err()
		})

		result, err := handler(context.Background(), "tools/call", nil)
		if err != nil {
			t.Fatalf("expected nil protocol error, got: %v", err)
		}

		toolResult, ok := result.(*mcp.CallToolResult)
		if !ok {
			t.Fatalf("expected *mcp.CallToolResult, got %T", result)
		}
		if !toolResult.IsError {
			t.Fatal("expected IsError to be true")
		}
		if len(toolResult.Content) == 0 {
			t.Fatal("expected non-empty content")
		}
		text, ok := toolResult.Content[0].(*mcp.TextContent)
		if !ok {
			t.Fatalf("expected *mcp.TextContent, got %T", toolResult.Content[0])
		}
		if !strings.HasPrefix(text.Text, "TIMEOUT:") {
			t.Fatalf("expected TIMEOUT prefix, got: %s", text.Text)
		}
		if !strings.Contains(text.Text, "Retry at most once") {
			t.Fatalf("expected retry hint, got: %s", text.Text)
		}
	})

	t.Run("no timeout when handler completes in time", func(t *testing.T) {
		middleware := toolTimeoutMiddleware(200 * time.Millisecond)

		handler := middleware(func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			time.Sleep(5 * time.Millisecond)
			return nil, ctx.Err()
		})

		_, err := handler(context.Background(), "tools/call", nil)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	t.Run("preserves successful result even if deadline fires after completion", func(t *testing.T) {
		middleware := toolTimeoutMiddleware(10 * time.Millisecond)

		successResult := &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "success data"}},
			IsError: false,
		}

		handler := middleware(func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			// Simulate handler that finishes just before deadline, then deadline fires
			time.Sleep(15 * time.Millisecond)
			// Return success even though context expired — the work completed
			return successResult, nil
		})

		result, err := handler(context.Background(), "tools/call", nil)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}

		toolResult, ok := result.(*mcp.CallToolResult)
		if !ok {
			t.Fatalf("expected *mcp.CallToolResult, got %T", result)
		}
		if toolResult.IsError {
			t.Fatal("expected successful result to be preserved, not replaced with timeout")
		}
		text := toolResult.Content[0].(*mcp.TextContent)
		if text.Text != "success data" {
			t.Fatalf("expected original content preserved, got: %s", text.Text)
		}
	})
}
