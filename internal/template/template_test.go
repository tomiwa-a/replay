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
			expected: " and yes",
		},
		{
			input:    "value: {{ count }}",
			data:     map[string]any{"count": 3},
			expected: "value: 3",
		},
		{
			input:    "flag: {{ enabled }}",
			data:     map[string]any{"enabled": true},
			expected: "flag: true",
		},
		{
			input:    "{{ a }}-{{ b }}",
			data:     map[string]any{"a": "x", "b": "y"},
			expected: "x-y",
		},
		{
			input:    "no vars here",
			data:     map[string]any{},
			expected: "no vars here",
		},
	}

	for _, tt := range tests {
		got := Render(tt.input, tt.data)
		if got != tt.expected {
			t.Errorf("Render(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestRenderNested(t *testing.T) {
	data := map[string]any{
		"user": map[string]any{
			"name": "alice",
			"addr": map[string]any{
				"city": "london",
			},
		},
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"{{ user.name }}", "alice"},
		{"{{ user.addr.city }}", "london"},
		{"{{ user.missing }}", ""},
		{"hi {{ user.name }} from {{ user.addr.city }}", "hi alice from london"},
	}

	for _, tt := range tests {
		got := Render(tt.input, data)
		if got != tt.expected {
			t.Errorf("Render(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestRenderWhitespace(t *testing.T) {
	data := map[string]any{"name": "world"}
	tests := []struct {
		input    string
		expected string
	}{
		{"{{name}}", "world"},
		{"{{ name }}", "world"},
		{"{{  name  }}", "world"},
	}

	for _, tt := range tests {
		got := Render(tt.input, data)
		if got != tt.expected {
			t.Errorf("Render(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
