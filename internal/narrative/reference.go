package narrative

import "fmt"

// ReferenceYesterday returns a light reference to yesterday's event, if any.
func ReferenceYesterday(m *Memory) string {
	y := m.GetYesterday()
	if y != nil {
		return fmt.Sprintf("Yesterday, %s", y.Detail)
	}
	return ""
}

// ReferenceLastOfType returns a light reference to the last event of a given type.
func ReferenceLastOfType(m *Memory, eventType string) string {
	e := m.GetLastOfType(eventType)
	if e != nil {
		return fmt.Sprintf("Last time this happened: %s", e.Detail)
	}
	return ""
}
