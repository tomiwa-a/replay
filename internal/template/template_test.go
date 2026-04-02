package template

import "testing"

func TestRender(t *testing.T) {
	tests := []struct {
		input    string
		data     map[string]any
		expected string
	}{
		{
			input:    "hello {{ name }}",
			data:     map[string]any{"name": "world"},
			expected: "hello world",
		},
		{
			input:    "api/v1/users/{{ id }}/posts",
			data:     map[string]any{"id": 42},
			expected: "api/v1/users/42/posts",
		},
		{
			input:    "{{ missing }} and {{ found }}",
			data:     map[string]any{"found": "yes"},
			expected: "{{ missing }} and yes",
		},
	}

	for _, tt := range tests {
		got := Render(tt.input, tt.data)
		if got != tt.expected {
			t.Errorf("Render(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
