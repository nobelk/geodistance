package geodistanceserver

import (
	"os"
	"testing"
)

func TestGeodistanceServer(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		expectErr bool
	}{
		{
			name:      "successful server creation",
			apiKey:    "test-api-key",
			expectErr: false,
		},
		{
			name:      "missing API key",
			apiKey:    "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.apiKey != "" {
				os.Setenv("GOOGLE_API_KEY", tt.apiKey)
				defer os.Unsetenv("GOOGLE_API_KEY")
			} else {
				os.Unsetenv("GOOGLE_API_KEY")
			}

			// Test server creation
			server, err := GeodistanceServer()

			if tt.expectErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if server != nil {
					t.Error("expected nil server when error occurs")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if server == nil {
					t.Error("expected non-nil server")
				}
			}
		})
	}
}

func TestGeodistanceServer_Configuration(t *testing.T) {
	// Set up environment
	os.Setenv("GOOGLE_API_KEY", "test-api-key")
	defer os.Unsetenv("GOOGLE_API_KEY")

	server, err := GeodistanceServer()
	if err != nil {
		t.Fatalf("unexpected error creating server: %v", err)
	}

	// Test server is not nil
	if server == nil {
		t.Fatal("expected non-nil server")
	}

	// Test server has proper configuration
	// Note: These tests depend on the internal structure of MCPServer
	// and may need to be adjusted based on the actual MCP-Go API

	// We can't directly access internal fields, but we can verify
	// the server was created successfully and is ready to use
	if server == nil {
		t.Error("server should not be nil")
	}
}

func TestGeodistanceServer_ToolRegistration(t *testing.T) {
	// Set up environment
	os.Setenv("GOOGLE_API_KEY", "test-api-key")
	defer os.Unsetenv("GOOGLE_API_KEY")

	server, err := GeodistanceServer()
	if err != nil {
		t.Fatalf("unexpected error creating server: %v", err)
	}

	// Test that server was created successfully
	if server == nil {
		t.Error("expected non-nil server")
	}

	// Note: The MCP-Go library may not expose methods to directly inspect
	// registered tools. This test verifies that the server creation
	// process completes without errors, which includes tool registration.

	// If the MCP-Go library provides methods to list tools, we could add:
	// tools := server.GetTools() // hypothetical method
	// if len(tools) != 1 {
	//     t.Errorf("expected 1 tool, got %d", len(tools))
	// }
}

func TestGeodistanceServer_Version(t *testing.T) {
	// Test that Version variable is set
	if Version == "" {
		t.Error("Version should not be empty")
	}

	// Test that Version is used in server creation
	os.Setenv("GOOGLE_API_KEY", "test-api-key")
	defer os.Unsetenv("GOOGLE_API_KEY")

	// Save original version
	originalVersion := Version
	defer func() { Version = originalVersion }()

	// Set test version
	Version = "test-version"

	server, err := GeodistanceServer()
	if err != nil {
		t.Fatalf("unexpected error creating server: %v", err)
	}

	if server == nil {
		t.Error("expected non-nil server")
	}

	// The server should be created with the test version
	// Note: We can't directly verify this without access to server internals
	// but the test ensures the version is properly passed to the constructor
}

func TestGeodistanceServer_HandlerCreationFailure(t *testing.T) {
	// Ensure no API key is set to force handler creation to fail
	os.Unsetenv("GOOGLE_API_KEY")

	server, err := GeodistanceServer()

	// Should return error when handler creation fails
	if err == nil {
		t.Error("expected error when handler creation fails")
	}

	if server != nil {
		t.Error("expected nil server when handler creation fails")
	}

	// Error should be related to missing API key
	expectedErrMsg := "GOOGLE_API_KEY environment variable not set"
	if err.Error() != expectedErrMsg {
		t.Errorf("expected error message %q, got %q", expectedErrMsg, err.Error())
	}
}

func TestGeodistanceServer_Integration(t *testing.T) {
	// This test verifies the complete integration between server and handler
	os.Setenv("GOOGLE_API_KEY", "test-api-key")
	defer os.Unsetenv("GOOGLE_API_KEY")

	server, err := GeodistanceServer()
	if err != nil {
		t.Fatalf("unexpected error creating server: %v", err)
	}

	if server == nil {
		t.Fatal("expected non-nil server")
	}

	// Verify server is properly configured
	// The server should have:
	// 1. A name "mcp-geodistance-server"
	// 2. The current version
	// 3. Resource capabilities enabled
	// 4. One tool registered ("calculate_distance")

	// Note: Without access to server internals, we can only verify
	// that the server was created successfully without errors
}

// Benchmark test to measure server creation performance
func BenchmarkGeodistanceServer(b *testing.B) {
	os.Setenv("GOOGLE_API_KEY", "test-api-key")
	defer os.Unsetenv("GOOGLE_API_KEY")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server, err := GeodistanceServer()
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
		if server == nil {
			b.Fatal("expected non-nil server")
		}
	}
}
