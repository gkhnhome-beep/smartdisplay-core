// Package contexthelp provides deterministic, action-oriented help for the current UI screen and state.
package contexthelp

// HelpRequest describes the current UI context for help generation.
type HelpRequest struct {
	Screen string
	State  map[string]interface{}
}

// HelpResponse is a short, action-oriented help message.
type HelpResponse struct {
	Message string `json:"message"`
}

// GenerateHelp returns a deterministic, 1–2 sentence, action-oriented help message for the given context.
func GenerateHelp(req HelpRequest) HelpResponse {
	switch req.Screen {
	case "home":
		return HelpResponse{Message: "Tap any icon to access its features. Swipe left or right to see more options."}
	case "alarm":
		if req.State["armed"] == true {
			return HelpResponse{Message: "To disarm, enter your code and tap Disarm. If you need help, tap the info icon."}
		}
		return HelpResponse{Message: "To arm the system, choose a mode and tap Arm. For details, tap the info icon."}
	case "guest":
		return HelpResponse{Message: "Add a guest by entering their name and tapping Add. Remove guests from the list below."}
	// ...add more screens as needed...
	default:
		return HelpResponse{Message: "Use the available buttons to perform actions. For more help, tap the info icon."}
	}
}
