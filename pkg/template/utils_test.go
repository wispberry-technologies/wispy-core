package template

import (
	"encoding/json"
	"reflect"
	"testing"

	"wispy-core/pkg/models"
)

func TestJSONFilter(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  string
	}{
		{
			name:  "string value",
			value: "hello",
			want:  `"hello"`,
		},
		{
			name: "map value",
			value: map[string]interface{}{
				"name": "John",
				"age":  30,
			},
			want: "{\n  \"age\": 30,\n  \"name\": \"John\"\n}",
		},
		{
			name:  "nil value",
			value: nil,
			want:  "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var typ reflect.Type
			if tt.value != nil {
				typ = reflect.TypeOf(tt.value)
			}
			result := JSONFilter(tt.value, typ, nil)

			// If result is a string, compare it with want
			if str, ok := result.(string); ok {
				// Normalize JSON for comparison
				var gotObj, wantObj interface{}
				if err := json.Unmarshal([]byte(str), &gotObj); err != nil {
					t.Errorf("failed to parse result JSON: %v", err)
					return
				}
				if err := json.Unmarshal([]byte(tt.want), &wantObj); err != nil {
					t.Errorf("failed to parse expected JSON: %v", err)
					return
				}

				gotJSON, _ := json.Marshal(gotObj)
				wantJSON, _ := json.Marshal(wantObj)

				if string(gotJSON) != string(wantJSON) {
					t.Errorf("expected %s, got %s", tt.want, str)
				}
			} else {
				t.Errorf("expected string result, got %T", result)
			}
		})
	}
}

func TestResolveFilterChain(t *testing.T) {
	// Create test context
	ctx := &models.TemplateContext{
		Data: map[string]interface{}{
			"Site": map[string]interface{}{
				"name":    "Test Site",
				"version": "1.0",
			},
		},
	}

	// Create filter map
	filters := GetDefaultFilters()

	tests := []struct {
		name    string
		chain   string
		wantNil bool
		wantErr bool
	}{
		{
			name:    "simple value",
			chain:   ".Site",
			wantNil: false,
		},
		{
			name:    "json filter",
			chain:   ".Site | json",
			wantNil: false,
		},
		{
			name:    "missing value",
			chain:   ".Missing",
			wantNil: true,
		},
		{
			name:    "unknown filter",
			chain:   ".Site | unknown",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, errs := ResolveFilterChain(tt.chain, ctx, filters)

			// Check for expected error
			if tt.wantErr && len(errs) == 0 {
				t.Error("expected error, got none")
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("unexpected errors: %v", errs)
			}

			// Check nil result
			if tt.wantNil && result != nil {
				t.Errorf("expected nil result, got %v", result)
			}
			if !tt.wantNil && result == nil {
				t.Error("expected non-nil result, got nil")
			}
		})
	}
}
