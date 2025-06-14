package tail

import (
	"reflect"
	"testing"
)

func TestGetColorValue(t *testing.T) {
	tests := []struct {
		name      string
		colorName string
		want      string
		exists    bool
	}{
		{
			name:      "standard tailwind color",
			colorName: "gray-950",
			want:      "var(--color-gray-950)",
			exists:    true,
		},
		{
			name:      "daisyUI primary color",
			colorName: "primary",
			want:      "var(--color-primary)",
			exists:    true,
		},
		{
			name:      "daisyUI base-100 color",
			colorName: "base-100",
			want:      "var(--color-base-100)",
			exists:    true,
		},
		{
			name:      "non-existent color",
			colorName: "nonexistent-color",
			want:      "",
			exists:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, exists := GetColorValue(tt.colorName)
			if exists != tt.exists {
				t.Errorf("GetColorValue() exists = %v, want %v", exists, tt.exists)
			}
			if got != tt.want {
				t.Errorf("GetColorValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandleColorClass(t *testing.T) {
	tests := []struct {
		name    string
		class   string
		want    []CSSProperty
		wantErr bool
	}{
		{
			name:  "bg-primary",
			class: "bg-primary",
			want: []CSSProperty{
				{"background-color", "var(--color-primary)"},
			},
			wantErr: false,
		},
		{
			name:  "text-base-100",
			class: "text-base-100",
			want: []CSSProperty{
				{"color", "var(--color-base-100)"},
			},
			wantErr: false,
		},
		{
			name:  "dark:text-primary",
			class: "dark:text-primary",
			want: []CSSProperty{
				{"@media (prefers-color-scheme: dark)", "color: var(--color-primary);"},
			},
			wantErr: false,
		},
		{
			name:  "border-error",
			class: "border-error",
			want: []CSSProperty{
				{"border-color", "var(--color-error)"},
			},
			wantErr: false,
		},
		{
			name:  "bg-primary/50",
			class: "bg-primary/50",
			want: []CSSProperty{
				{"background-color", "var(--color-primary) / 0.50"},
			},
			wantErr: false,
		},
		{
			name:  "dark:bg-primary/50",
			class: "dark:bg-primary/50",
			want: []CSSProperty{
				{"@media (prefers-color-scheme: dark)", "background-color: var(--color-primary) / 0.50;"},
			},
			wantErr: false,
		},
		{
			name:    "invalid-class",
			class:   "invalid",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HandleColorClass(tt.class)
			if (err != nil) != tt.wantErr {
				t.Errorf("HandleColorClass() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HandleColorClass() = %v, want %v", got, tt.want)
			}
		})
	}
}
