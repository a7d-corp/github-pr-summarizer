package analysis

import (
	"fmt"
	"time"
)

// FormatAge renders how long ago createdAt was, relative to now, as a short
// duration string: "42m", "5h", "3d", or "2mo".
func FormatAge(createdAt, now time.Time) string {
	d := now.Sub(createdAt)
	if d < 0 {
		d = 0
	}

	switch {
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	default:
		return fmt.Sprintf("%dmo", int(d.Hours()/24/30))
	}
}
