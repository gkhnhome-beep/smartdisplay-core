package i18n

import (
	"testing"
)

func TestInit(t *testing.T) {
	if err := Init("en"); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if !IsInitialized() {
		t.Error("Expected i18n to be initialized")
	}

	if lang := GetLang(); lang != "en" {
		t.Errorf("Expected language 'en', got %s", lang)
	}
}

func TestSetLang(t *testing.T) {
	Init("en")

	SetLang("tr")
	if lang := GetLang(); lang != "tr" {
		t.Errorf("Expected language 'tr', got %s", lang)
	}

	// Test fallback to default when unknown language
	SetLang("unknown")
	if lang := GetLang(); lang != "en" {
		t.Errorf("Expected fallback to 'en', got %s", lang)
	}
}

func TestTranslation(t *testing.T) {
	Init("en")

	// Test existing key
	text := T("ai.system_normal")
	if text == "" || text == "ai.system_normal" {
		t.Errorf("Expected translation for 'ai.system_normal', got %s", text)
	}

	// Test missing key (should return key itself)
	text = T("nonexistent.key")
	if text != "nonexistent.key" {
		t.Errorf("Expected 'nonexistent.key' for missing key, got %s", text)
	}
}

func TestFallbackToEnglish(t *testing.T) {
	Init("tr")

	// Test key that exists in English but might not in Turkish
	// Even if it exists in Turkish, we're testing the fallback mechanism
	text := T("ai.system_normal")
	if text == "" || text == "ai.system_normal" {
		t.Errorf("Expected translation (with fallback), got %s", text)
	}
}

func TestGetAvailableLanguages(t *testing.T) {
	Init("en")

	langs := GetAvailableLanguages()
	if len(langs) == 0 {
		t.Error("Expected at least one language to be available")
	}

	hasEnglish := false
	for _, lang := range langs {
		if lang == "en" {
			hasEnglish = true
			break
		}
	}

	if !hasEnglish {
		t.Error("Expected 'en' to be in available languages")
	}
}

func TestThreadSafety(t *testing.T) {
	Init("en")

	// Concurrent reads and writes
	done := make(chan bool, 10)

	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				SetLang("en")
				T("ai.system_normal")
			}
			done <- true
		}()
	}

	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				SetLang("tr")
				T("ai.system_normal")
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestUninitializedBehavior(t *testing.T) {
	// Reset by creating new instance (in real scenario, don't do this)
	// This test verifies fallback when not initialized
	mu.Lock()
	initialized = false
	mu.Unlock()

	// Should return key when not initialized
	text := T("ai.system_normal")
	if text != "ai.system_normal" {
		t.Errorf("Expected key 'ai.system_normal' when uninitialized, got %s", text)
	}

	// Re-initialize for other tests
	Init("en")
}
