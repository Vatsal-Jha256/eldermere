package storage

import "testing"

func TestNormalizeDisplayName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "blank", input: "   ", want: "Wanderer"},
		{name: "collapses whitespace", input: "  Lady   of   the Lake  ", want: "Lady of the Lake"},
		{name: "limits length", input: "123456789012345678901234567890", want: "1234567890123456789012345678"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := normalizeDisplayName(test.input); got != test.want {
				t.Fatalf("normalizeDisplayName(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}
