package ui

// upstreamChoiceForPreset keeps GUI bundle keys aligned with the interactive
// GoECS menu. The upstream runner uses this value to select both scheduling
// and the fixed Chinese preset speed profile.
func upstreamChoiceForPreset(key string) string {
	switch key {
	case "full":
		return "1"
	case "full_concurrent":
		return "2"
	case "minimal":
		return "3"
	case "standard":
		return "4"
	case "network_focus":
		return "5"
	case "unlock_focus":
		return "6"
	case "network_only":
		return "7"
	case "unlock_only":
		return "8"
	case "hardware_only":
		return "9"
	case "ip_quality":
		return "10"
	case "route_only":
		return "11"
	default:
		return ""
	}
}
