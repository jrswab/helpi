package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jrswab/helpi/internal/llm"
)

func TestNewManager_EmptyPathUsesDefault(t *testing.T) {
	mgr, err := NewManager("", 10)
	if err != nil {
		t.Fatalf("NewManager() returned error: %v", err)
	}

	m := mgr.(*manager)
	if m.path != "./data/sessions" {
		t.Errorf("expected path to be ./data/sessions, got %s", m.path)
	}
}

func TestNewManager_MaxMessagesZeroUsesDefault(t *testing.T) {
	mgr, err := NewManager(t.TempDir(), 0)
	if err != nil {
		t.Fatalf("NewManager() returned error: %v", err)
	}

	m := mgr.(*manager)
	if m.maxMessages != 50 {
		t.Errorf("expected maxMessages to be 50, got %d", m.maxMessages)
	}
}

func TestNewManager_InvalidDirectoryReturnsError(t *testing.T) {
	nonWritable := "/root/nonexistent/invalid/path"
	_, err := NewManager(nonWritable, 10)
	if err == nil {
		t.Error("expected error for non-writable directory")
	}
}

func TestGet_NoSessionFile(t *testing.T) {
	mgr, err := NewManager(t.TempDir(), 10)
	if err != nil {
		t.Fatalf("NewManager() returned error: %v", err)
	}

	msgs, err := mgr.Get(12345)
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}
	if len(msgs) != 0 {
		t.Errorf("expected empty slice, got %d messages", len(msgs))
	}
}

func TestGet_ValidJSON(t *testing.T) {
	dir := t.TempDir()
	mgr, err := NewManager(dir, 10)
	if err != nil {
		t.Fatalf("NewManager() returned error: %v", err)
	}

	expected := []llm.Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there"},
	}
	data, _ := json.Marshal(expected)
	sessionPath := filepath.Join(dir, "12345.json")
	if err := os.WriteFile(sessionPath, data, 0644); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}

	msgs, err := mgr.Get(12345)
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}
	if len(msgs) != 2 {
		t.Errorf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[0].Role != "user" || msgs[0].Content != "Hello" {
		t.Error("first message content mismatch")
	}
	if msgs[1].Role != "assistant" || msgs[1].Content != "Hi there" {
		t.Error("second message content mismatch")
	}
}

func TestGet_CorruptedJSON(t *testing.T) {
	dir := t.TempDir()
	mgr, err := NewManager(dir, 10)
	if err != nil {
		t.Fatalf("NewManager() returned error: %v", err)
	}

	sessionPath := filepath.Join(dir, "12345.json")
	if err := os.WriteFile(sessionPath, []byte("invalid json{"), 0644); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}

	_, err = mgr.Get(12345)
	if err == nil {
		t.Error("expected error for corrupted JSON")
	}
}

func TestSave_NewSessionCreatesFile(t *testing.T) {
	dir := t.TempDir()
	mgr, err := NewManager(dir, 10)
	if err != nil {
		t.Fatalf("NewManager() returned error: %v", err)
	}

	messages := []llm.Message{
		{Role: "user", Content: "Test message"},
	}
	if err := mgr.Save(12345, messages); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	sessionPath := filepath.Join(dir, "12345.json")
	if _, err := os.ReadFile(sessionPath); err != nil {
		t.Errorf("expected session file to exist: %v", err)
	}
}

func TestSave_ExceedsMaxMessagesTruncates(t *testing.T) {
	dir := t.TempDir()
	mgr, err := NewManager(dir, 3)
	if err != nil {
		t.Fatalf("NewManager() returned error: %v", err)
	}

	messages := []llm.Message{
		{Role: "user", Content: "msg1"},
		{Role: "assistant", Content: "msg2"},
		{Role: "user", Content: "msg3"},
		{Role: "assistant", Content: "msg4"},
		{Role: "user", Content: "msg5"},
	}
	if err := mgr.Save(12345, messages); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	msgs, err := mgr.Get(12345)
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}
	if len(msgs) != 3 {
		t.Errorf("expected 3 messages after truncation, got %d", len(msgs))
	}
	if msgs[0].Content != "msg3" {
		t.Errorf("expected first message to be msg3, got %s", msgs[0].Content)
	}
	if msgs[2].Content != "msg5" {
		t.Errorf("expected last message to be msg5, got %s", msgs[2].Content)
	}
}

func TestDelete_ExistingFileRemovesIt(t *testing.T) {
	dir := t.TempDir()
	mgr, err := NewManager(dir, 10)
	if err != nil {
		t.Fatalf("NewManager() returned error: %v", err)
	}

	messages := []llm.Message{{Role: "user", Content: "Test"}}
	if err := mgr.Save(12345, messages); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	if err := mgr.Delete(12345); err != nil {
		t.Fatalf("Delete() returned error: %v", err)
	}

	sessionPath := filepath.Join(dir, "12345.json")
	if _, err := os.ReadFile(sessionPath); err == nil {
		t.Error("expected session file to be deleted")
	}
}

func TestDelete_NonExistentFileReturnsNil(t *testing.T) {
	mgr, err := NewManager(t.TempDir(), 10)
	if err != nil {
		t.Fatalf("NewManager() returned error: %v", err)
	}

	err = mgr.Delete(99999)
	if err != nil {
		t.Errorf("Delete() returned error for non-existent file: %v", err)
	}
}
