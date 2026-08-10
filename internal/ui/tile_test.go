package ui

import "testing"

func TestRangeColorFor(t *testing.T) {
	r := Range{Min: 0, Max: 100, WarnLow: 20, WarnHigh: 80, DangerLow: 10, DangerHigh: 90}
	theme := Dark

	cases := []struct {
		value float32
		want  string
	}{
		{5, "danger"},   // below DangerLow
		{15, "warning"}, // between DangerLow and WarnLow
		{50, "good"},    // inside the safe band
		{85, "warning"}, // between WarnHigh and DangerHigh
		{95, "danger"},  // above DangerHigh
	}

	for _, c := range cases {
		got := r.colorFor(c.value, theme)
		var want = theme.Good
		switch c.want {
		case "danger":
			want = theme.Danger
		case "warning":
			want = theme.Warning
		}
		if got != want {
			t.Errorf("colorFor(%v) = %v, want %s (%v)", c.value, got, c.want, want)
		}
	}
}

func TestRangeFractionClamps(t *testing.T) {
	r := Range{Min: 0, Max: 100}

	if f := r.fraction(-50); f != 0 {
		t.Errorf("fraction(-50) = %v, want 0", f)
	}
	if f := r.fraction(150); f != 1 {
		t.Errorf("fraction(150) = %v, want 1", f)
	}
	if f := r.fraction(25); f != 0.25 {
		t.Errorf("fraction(25) = %v, want 0.25", f)
	}
}
