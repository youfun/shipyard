package deploy

import (
	"testing"
)

func TestFormatEnvValue(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  string
	}{
		{
			name:  "Simple string",
			value: "hello",
			want:  "hello",
		},
		{
			name:  "Integer",
			value: 4000,
			want:  "4000",
		},
		{
			name:  "Boolean true",
			value: true,
			want:  "true",
		},
		{
			name:  "Boolean false",
			value: false,
			want:  "false",
		},
		{
			name:  "String slice",
			value: []string{"a", "b", "c"},
			want:  "a,b,c",
		},
		{
			name:  "Empty string slice",
			value: []string{},
			want:  "",
		},
		{
			name:  "Single element string slice",
			value: []string{"only"},
			want:  "only",
		},
		{
			name:  "Interface slice",
			value: []interface{}{"x", "y", "z"},
			want:  "x,y,z",
		},
		{
			name:  "Mixed interface slice",
			value: []interface{}{"a", 1, true},
			want:  "a,1,true",
		},
		{
			name:  "Empty interface slice",
			value: []interface{}{},
			want:  "",
		},
		{
			name:  "Float value",
			value: 3.14,
			want:  "3.14",
		},
		{
			name:  "Domains as string slice",
			value: []string{"example.com", "www.example.com"},
			want:  "example.com,www.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatEnvValue(tt.value)
			if got != tt.want {
				t.Errorf("formatEnvValue(%v) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}
