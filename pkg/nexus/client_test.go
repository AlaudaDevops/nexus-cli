package nexus

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		username string
		password string
	}{
		{
			name:     "create client with valid credentials",
			baseURL:  "http://localhost:8081",
			username: "admin",
			password: "admin123",
		},
		{
			name:     "create client with trailing slash in URL",
			baseURL:  "http://localhost:8081/",
			username: "admin",
			password: "admin123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.baseURL, tt.username, tt.password)

			if client == nil {
				t.Fatal("NewClient() returned nil")
			}

			if client.username != tt.username {
				t.Errorf("NewClient() username = %v, want %v", client.username, tt.username)
			}

			if client.password != tt.password {
				t.Errorf("NewClient() password = %v, want %v", client.password, tt.password)
			}

			// URL should have trailing slash removed
			expectedURL := "http://localhost:8081"
			if client.baseURL != expectedURL {
				t.Errorf("NewClient() baseURL = %v, want %v", client.baseURL, expectedURL)
			}

			if client.httpClient == nil {
				t.Error("NewClient() httpClient is nil")
			}
		})
	}
}
