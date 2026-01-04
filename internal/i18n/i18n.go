// Package i18n provides localization (internationalization) support for smartdisplay-core.
// It supports loading language files from configs/lang/ directory and provides
// thread-safe translation lookups with fallback to English.
package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

var (
	mu           sync.RWMutex
	currentLang  string
	translations map[string]map[string]string // lang -> key -> translation
	defaultLang  string
	initialized  bool
)

// Init initializes the i18n system with a default language.
// It loads available language files from configs/lang/ directory.
// If language files are missing or broken, it logs a warning and continues with fallback.
func Init(lang string) error {
	mu.Lock()
	defer mu.Unlock()

	if lang == "" {
		lang = "en"
	}

	defaultLang = lang
	currentLang = lang
	translations = make(map[string]map[string]string)

	// Load available language files
	langDir := "configs/lang"
	supportedLangs := []string{"en", "tr"}

	for _, l := range supportedLangs {
		path := filepath.Join(langDir, l+".json")
		if err := loadLanguageFile(l, path); err != nil {
			// Log warning but continue - fallback will handle missing translations
			println("i18n: warning: failed to load language file " + path + ": " + err.Error())
		}
	}

	// Ensure at least English is available (even if empty)
	if translations["en"] == nil {
		translations["en"] = make(map[string]string)
	}

	initialized = true
	return nil
}

// loadLanguageFile loads a JSON language file into the translations map.
func loadLanguageFile(lang, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var langMap map[string]string
	if err := json.Unmarshal(data, &langMap); err != nil {
		return err
	}

	translations[lang] = langMap
	return nil
}

// SetLang changes the current language at runtime.
// If the language is not loaded, it falls back to the default language.
func SetLang(lang string) {
	mu.Lock()
	defer mu.Unlock()

	if translations[lang] != nil {
		currentLang = lang
	} else {
		// Fallback to default if language not available
		currentLang = defaultLang
		println("i18n: warning: language '" + lang + "' not available, using '" + defaultLang + "'")
	}
}

// GetLang returns the currently active language.
func GetLang() string {
	mu.RLock()
	defer mu.RUnlock()
	return currentLang
}

// T translates a key to the current language.
// If the key is not found in the current language, it falls back to English.
// If the key is not found in English either, it returns the key itself.
func T(key string) string {
	mu.RLock()
	defer mu.RUnlock()

	if !initialized {
		return key
	}

	// Try current language
	if trans, ok := translations[currentLang][key]; ok {
		return trans
	}

	// Fallback to English
	if currentLang != "en" {
		if trans, ok := translations["en"][key]; ok {
			return trans
		}
	}

	// Fallback to key itself
	return key
}

// IsInitialized returns whether the i18n system has been initialized.
func IsInitialized() bool {
	mu.RLock()
	defer mu.RUnlock()
	return initialized
}

// GetAvailableLanguages returns a list of loaded languages.
func GetAvailableLanguages() []string {
	mu.RLock()
	defer mu.RUnlock()

	langs := make([]string, 0, len(translations))
	for lang := range translations {
		langs = append(langs, lang)
	}
	return langs
}
