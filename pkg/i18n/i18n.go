package i18n

var (
	currentLocale = "en"
	messages      = map[string]map[string]string{
		"en": enMessages,
		"zh": zhMessages,
	}
)

// SetLocale sets the current locale
func SetLocale(locale string) {
	currentLocale = locale
}

// GetLocale returns the current locale
func GetLocale() string {
	return currentLocale
}

// Get returns a localized message by key
func Get(key string) string {
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
	currentLocale = "en"
}
