package utils

import "testing"

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "spaces and punctuation", input: "  Main Branch, Central! ", want: "main-branch-central"},
		{name: "repeated separators", input: "branch---one___two", want: "branch-one-two"},
		{name: "unicode letters and digits", input: "Café 2", want: "café-2"},
		{name: "empty input", input: "", want: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := GenerateSlug(test.input); got != test.want {
				t.Errorf("GenerateSlug(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}
