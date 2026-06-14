package i18n

import (
	"sync"
)

var (
	currentLocale = "en"
	mu            sync.RWMutex
	messages      = map[string]map[string]string{
		"en": enMessages,
		"zh": zhMessages,
	}
)

// SetLocale sets the current locale
func SetLocale(locale string) {
	mu.Lock()
	defer mu.Unlock()
	currentLocale = locale
}

// GetLocale returns the current locale
func GetLocale() string {
	mu.RLock()
	defer mu.RUnlock()
	return currentLocale
}

// Get returns a localized message by key
func Get(key string) string {
	mu.RLock()
	defer mu.RUnlock()

	if localeMessages, ok := messages[currentLocale]; ok {
		if msg, ok := localeMessages[key]; ok {
			return msg
		}
	}

	if msg, ok := messages["en"][key]; ok {
		return msg
	}

	return key
}

// GetDescription returns the localized description from zh/en strings
func GetDescription(zh, en string) string {
	mu.RLock()
	defer mu.RUnlock()

	switch currentLocale {
	case "zh":
		return zh
	default:
		return en
	}
}

// SupportedLocales returns the list of supported locales
func SupportedLocales() []string {
	locales := make([]string, 0, len(messages))
	for locale := range messages {
		locales = append(locales, locale)
	}
	return locales
}

// Reset resets the i18n manager to default state
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	currentLocale = "en"
}
