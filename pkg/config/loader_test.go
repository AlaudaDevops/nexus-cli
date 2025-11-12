package config

import (
	"os"
	"testing"
)

func TestGetNexusCredentials(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		wantURL     string
		wantUser    string
		wantPass    string
		wantErr     bool
		errContains string
	}{
		{
			name: "all environment variables set",
			envVars: map[string]string{
				"NEXUS_URL":      "http://localhost:8081",
				"NEXUS_USERNAME": "admin",
				"NEXUS_PASSWORD": "admin123",
			},
			wantURL:  "http://localhost:8081",
			wantUser: "admin",
			wantPass: "admin123",
			wantErr:  false,
		},
		{
			name: "missing NEXUS_URL",
			envVars: map[string]string{
				"NEXUS_USERNAME": "admin",
				"NEXUS_PASSWORD": "admin123",
			},
			wantErr:     true,
			errContains: "NEXUS_URL",
		},
		{
			name: "missing NEXUS_USERNAME",
			envVars: map[string]string{
				"NEXUS_URL":      "http://localhost:8081",
				"NEXUS_PASSWORD": "admin123",
			},
			wantErr:     true,
			errContains: "NEXUS_USERNAME",
		},
		{
			name: "missing NEXUS_PASSWORD",
			envVars: map[string]string{
				"NEXUS_URL":      "http://localhost:8081",
				"NEXUS_USERNAME": "admin",
			},
			wantErr:     true,
			errContains: "NEXUS_PASSWORD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variables
			os.Unsetenv("NEXUS_URL")
			os.Unsetenv("NEXUS_USERNAME")
			os.Unsetenv("NEXUS_PASSWORD")

			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			url, user, pass, err := GetNexusCredentials()

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetNexusCredentials() expected error containing %q, got nil", tt.errContains)
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("GetNexusCredentials() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("GetNexusCredentials() unexpected error = %v", err)
				return
			}

			if url != tt.wantURL {
				t.Errorf("GetNexusCredentials() url = %v, want %v", url, tt.wantURL)
			}
			if user != tt.wantUser {
				t.Errorf("GetNexusCredentials() user = %v, want %v", user, tt.wantUser)
			}
			if pass != tt.wantPass {
				t.Errorf("GetNexusCredentials() pass = %v, want %v", pass, tt.wantPass)
			}

			// Clean up
			for k := range tt.envVars {
				os.Unsetenv(k)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsInside(s, substr)))
}

func containsInside(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
