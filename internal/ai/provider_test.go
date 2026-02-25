// Copyright (c) 2026 Madalin Gabriel Ignisca <hi@madalin.me>
// Copyright (c) 2026 Vlah Software House SRL <contact@vlah.sh>
// All rights reserved. See LICENSE for details.

package ai

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestGeminiLive tests the Gemini provider against the real API.
// Skipped if GEMINI_API_KEY is not set.
func TestGeminiLive(t *testing.T) {
	key := os.Getenv("GEMINI_API_KEY")
	if key == "" {
		t.Skip("GEMINI_API_KEY not set")
	}

	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-3.1-pro-preview"
	}

	reg := NewRegistry("gemini", map[string]ProviderConfig{
		"gemini": {APIKey: key, Model: model},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := reg.Generate(ctx, "Reply in exactly one short sentence.", "What is 2+2?")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if result == "" {
		t.Fatal("Generate returned empty string")
	}

	t.Logf("Gemini response: %s", result)
}

// TestClaudeLive tests the Claude provider against the real API.
// Skipped if CLAUDE_API_KEY is not set.
func TestClaudeLive(t *testing.T) {
	key := os.Getenv("CLAUDE_API_KEY")
	if key == "" {
		t.Skip("CLAUDE_API_KEY not set")
	}

	model := os.Getenv("CLAUDE_MODEL")
	if model == "" {
		model = "claude-sonnet-4-6"
	}

	reg := NewRegistry("claude", map[string]ProviderConfig{
		"claude": {APIKey: key, Model: model},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := reg.Generate(ctx, "Reply in exactly one short sentence.", "What is 2+2?")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if result == "" {
		t.Fatal("Generate returned empty string")
	}

	t.Logf("Claude response: %s", result)
}

// TestOpenAILive tests the OpenAI provider against the real API.
// Skipped if OPENAI_API_KEY is not set.
func TestOpenAILive(t *testing.T) {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-4o"
	}

	reg := NewRegistry("openai", map[string]ProviderConfig{
		"openai": {APIKey: key, Model: model},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := reg.Generate(ctx, "Reply in exactly one short sentence.", "What is 2+2?")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if result == "" {
		t.Fatal("Generate returned empty string")
	}

	t.Logf("OpenAI response: %s", result)
}

// TestMistralLive tests the Mistral provider against the real API.
// Skipped if MISTRAL_API_KEY is not set.
func TestMistralLive(t *testing.T) {
	key := os.Getenv("MISTRAL_API_KEY")
	if key == "" {
		t.Skip("MISTRAL_API_KEY not set")
	}

	model := os.Getenv("MISTRAL_MODEL")
	if model == "" {
		model = "mistral-large-latest"
	}

	reg := NewRegistry("mistral", map[string]ProviderConfig{
		"mistral": {APIKey: key, Model: model},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := reg.Generate(ctx, "Reply in exactly one short sentence.", "What is 2+2?")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if result == "" {
		t.Fatal("Generate returned empty string")
	}

	t.Logf("Mistral response: %s", result)
}

// TestRegistryBasics tests registry provider management without API calls.
func TestRegistryBasics(t *testing.T) {
	reg := NewRegistry("gemini", map[string]ProviderConfig{
		"openai":  {APIKey: "test-key", Model: "gpt-4o"},
		"gemini":  {APIKey: "test-key", Model: "gemini-pro"},
		"claude":  {APIKey: "", Model: "claude-sonnet"}, // No key — should be skipped.
		"mistral": {APIKey: "test-key", Model: "mistral-large"},
	})

	if reg.ActiveName() != "gemini" {
		t.Errorf("expected active=gemini, got %s", reg.ActiveName())
	}

	if reg.HasProvider("claude") {
		t.Error("claude should not be available (no API key)")
	}

	available := reg.Available()
	if len(available) != 3 {
		t.Errorf("expected 3 available providers, got %d: %v", len(available), available)
	}

	if err := reg.SetActive("openai"); err != nil {
		t.Errorf("SetActive(openai) failed: %v", err)
	}
	if reg.ActiveName() != "openai" {
		t.Errorf("expected active=openai after switch, got %s", reg.ActiveName())
	}

	if err := reg.SetActive("claude"); err == nil {
		t.Error("SetActive(claude) should fail (no API key)")
	}
}
