package pipeline

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestResponseCache_PutGet(t *testing.T) {
	dir := t.TempDir()
	c := newResponseCache(dir, 60)

	data := []byte(`{"products":[]}`)
	c.put("TestShop", "smartphones_0", 0, data)

	got, ok := c.get("TestShop", "smartphones_0", 0)
	if !ok {
		t.Fatal("expected cache hit")
	}
	if string(got) != string(data) {
		t.Errorf("got %q, want %q", got, data)
	}
}

func TestResponseCache_Miss(t *testing.T) {
	dir := t.TempDir()
	c := newResponseCache(dir, 60)

	_, ok := c.get("TestShop", "missing", 0)
	if ok {
		t.Error("expected cache miss")
	}
}

func TestResponseCache_Expired(t *testing.T) {
	dir := t.TempDir()
	c := newResponseCache(dir, 1) // 1 minute TTL

	data := []byte(`test`)
	c.put("TestShop", "cat", 0, data)

	// Backdate the file modification time
	path := filepath.Join(dir, "testshop", "cat.1.bin")
	past := time.Now().Add(-2 * time.Minute)
	if err := os.Chtimes(path, past, past); err != nil {
		t.Fatal(err)
	}

	_, ok := c.get("TestShop", "cat", 0)
	if ok {
		t.Error("expected cache miss for expired entry")
	}
}

func TestResponseCache_NilSafe(t *testing.T) {
	var c *responseCache

	data, ok := c.get("shop", "cat", 0)
	if ok || data != nil {
		t.Error("nil cache get should return nil, false")
	}

	// Should not panic
	c.put("shop", "cat", 0, []byte("test"))
}

func TestResponseCache_DisabledWhenEmpty(t *testing.T) {
	c := newResponseCache("", 60)
	if c != nil {
		t.Error("empty dir should return nil cache")
	}

	c = newResponseCache("/tmp/test", 0)
	if c != nil {
		t.Error("zero TTL should return nil cache")
	}
}
