package analysis

import (
	"testing"
	"time"
)

func TestFormatAge(t *testing.T) {
	now := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		createdAt time.Time
		want      string
	}{
		{"minutes", now.Add(-42 * time.Minute), "42m"},
		{"just now", now, "0m"},
		{"hours", now.Add(-5 * time.Hour), "5h"},
		{"just under a day", now.Add(-23 * time.Hour), "23h"},
		{"days", now.Add(-3 * 24 * time.Hour), "3d"},
		{"just under a month", now.Add(-29 * 24 * time.Hour), "29d"},
		{"months", now.Add(-60 * 24 * time.Hour), "2mo"},
		{"future createdAt clamps to 0m", now.Add(1 * time.Hour), "0m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatAge(tt.createdAt, now); got != tt.want {
				t.Errorf("FormatAge() = %q, want %q", got, tt.want)
			}
		})
	}
}
