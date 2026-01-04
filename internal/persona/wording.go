package persona

// ToneForPersona returns the tone for a given persona.
func ToneForPersona(p PersonaType) string {
	switch p {
	case PersonaOwner:
		return "direct"
	case PersonaAdult:
		return "respectful"
	case PersonaChild:
		return "gentle"
	case PersonaGuest:
		return "welcoming"
	default:
		return "neutral"
	}
}

// WordingForPersona returns a wording variant for a given persona and message key.
func WordingForPersona(p PersonaType, key string) string {
	switch p {
	case PersonaOwner:
		if key == "alarm_help" {
			return "To arm or disarm, use your code."
		}
	case PersonaAdult:
		if key == "alarm_help" {
			return "Enter your code to arm or disarm the system."
		}
	case PersonaChild:
		if key == "alarm_help" {
			return "Ask an adult for help with the alarm."
		}
	case PersonaGuest:
		if key == "alarm_help" {
			return "Please ask your host if you need to use the alarm."
		}
	}
	// Default fallback
	return "For help, tap the info icon."
}
