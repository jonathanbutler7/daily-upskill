package config

import (
	"reflect"
	"testing"
	"time"
)

// TestNewDefaultAppConfig exists to ensure that the default config doesn't change by accident
func TestNewDefaultAppConfig(t *testing.T) {
	tests := []struct {
		name string
		want *AppConfig
	}{
		{
			name: "ensure default config",
			want: &AppConfig{
				TestString:   "test-string",
				TestInt:      3,
				TestBool:     true,
				TestAddress:  "google.com:8080",
				TestDuration: 1 * time.Second,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDefaultAppConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDefaultAppConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
