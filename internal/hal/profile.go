package hal

// HardwareProfile defines which HAL devices are required/optional for a given profile
// Profiles: minimal, standard, full

type HardwareProfile string

const (
	ProfileMinimal  HardwareProfile = "minimal"
	ProfileStandard HardwareProfile = "standard"
	ProfileFull     HardwareProfile = "full"
)

type ProfileSpec struct {
	Required []string // Device types required
	Optional []string // Device types optional
}

var HardwareProfiles = map[HardwareProfile]ProfileSpec{
	ProfileMinimal: {
		Required: []string{"fan"},
		Optional: []string{"rfid", "rf433", "rgb_led"},
	},
	ProfileStandard: {
		Required: []string{"fan", "rfid", "rgb_led"},
		Optional: []string{"rf433"},
	},
	ProfileFull: {
		Required: []string{"fan", "rfid", "rf433", "rgb_led"},
		Optional: []string{},
	},
}
