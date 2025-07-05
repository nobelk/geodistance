package geodistanceserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// MockHTTPClient implements HTTPClient interface for testing
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

// Helper function to create mock response
func createMockResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

// Helper function to create valid API response
func createValidAPIResponse() string {
	response := ResponseBody{
		Routes: []Route{
			{
				DistanceMeters: 1000,
				Duration:       "5m",
				RouteLabels:    []string{"DEFAULT_ROUTE"},
			},
		},
	}
	data, _ := json.Marshal(response)
	return string(data)
}

func TestNewGeodistanceHandler(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		expectErr bool
	}{
		{
			name:      "valid API key",
			apiKey:    "test-api-key",
			expectErr: false,
		},
		{
			name:      "empty API key",
			apiKey:    "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.apiKey != "" {
				os.Setenv("GOOGLE_API_KEY", tt.apiKey)
				defer os.Unsetenv("GOOGLE_API_KEY")
			} else {
				os.Unsetenv("GOOGLE_API_KEY")
			}

			handler, err := NewGeodistanceHandler()

			if tt.expectErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if handler != nil {
					t.Error("expected nil handler when error occurs")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if handler == nil {
					t.Error("expected non-nil handler")
				}
				if handler.apiKey != tt.apiKey {
					t.Errorf("expected API key %s, got %s", tt.apiKey, handler.apiKey)
				}
			}
		})
	}
}

func TestNewGeodistanceHandlerWithClient(t *testing.T) {
	mockClient := &MockHTTPClient{}

	os.Setenv("GOOGLE_API_KEY", "test-key")
	defer os.Unsetenv("GOOGLE_API_KEY")

	handler, err := NewGeodistanceHandlerWithClient(mockClient)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if handler == nil {
		t.Error("expected non-nil handler")
	}
	if handler.client != mockClient {
		t.Error("expected mock client to be set")
	}
}

func TestGeodistanceHandler_validateAddresses(t *testing.T) {
	handler := &GeodistanceHandler{}

	tests := []struct {
		name        string
		origin      string
		destination string
		expectErr   bool
	}{
		{
			name:        "valid addresses",
			origin:      "New York",
			destination: "Los Angeles",
			expectErr:   false,
		},
		{
			name:        "empty origin",
			origin:      "",
			destination: "Los Angeles",
			expectErr:   true,
		},
		{
			name:        "empty destination",
			origin:      "New York",
			destination: "",
			expectErr:   true,
		},
		{
			name:        "both empty",
			origin:      "",
			destination: "",
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.validateAddresses(tt.origin, tt.destination)

			if tt.expectErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestGeodistanceHandler_buildRequestBody(t *testing.T) {
	handler := &GeodistanceHandler{}

	origins := []Origin{{Address: "New York"}}
	destinations := []Destination{{Address: "Los Angeles"}}

	body := handler.buildRequestBody(origins, destinations)

	if body == nil {
		t.Error("expected non-nil request body")
	}
	if len(body.Origins) != 1 || body.Origins[0].Address != "New York" {
		t.Error("origins not set correctly")
	}
	if len(body.Destinations) != 1 || body.Destinations[0].Address != "Los Angeles" {
		t.Error("destinations not set correctly")
	}
	if body.TravelMode != "DRIVE" {
		t.Error("travel mode not set correctly")
	}
	if body.RoutingPreference != "TRAFFIC_AWARE" {
		t.Error("routing preference not set correctly")
	}
}

func TestGeodistanceHandler_createRequest(t *testing.T) {
	handler := &GeodistanceHandler{apiKey: "test-key"}
	ctx := context.Background()

	body := &RequestBody{
		Origins:      []Origin{{Address: "New York"}},
		Destinations: []Destination{{Address: "Los Angeles"}},
		TravelMode:   "DRIVE",
	}

	req, err := handler.createRequest(ctx, body)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if req == nil {
		t.Error("expected non-nil request")
	}
	if req.Method != "POST" {
		t.Errorf("expected POST method, got %s", req.Method)
	}
	if req.Header.Get("Content-Type") != "application/json" {
		t.Error("content type not set correctly")
	}
	if req.Header.Get("X-Goog-Api-Key") != "test-key" {
		t.Error("API key header not set correctly")
	}
}

func TestGeodistanceHandler_processResponse(t *testing.T) {
	handler := &GeodistanceHandler{}

	tests := []struct {
		name       string
		statusCode int
		body       string
		expectErr  bool
	}{
		{
			name:       "successful response",
			statusCode: http.StatusOK,
			body:       createValidAPIResponse(),
			expectErr:  false,
		},
		{
			name:       "API error",
			statusCode: http.StatusBadRequest,
			body:       `{"error": "Invalid request"}`,
			expectErr:  true,
		},
		{
			name:       "invalid JSON",
			statusCode: http.StatusOK,
			body:       "invalid json",
			expectErr:  true,
		},
		{
			name:       "no routes",
			statusCode: http.StatusOK,
			body:       `{"routes": []}`,
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := createMockResponse(tt.statusCode, tt.body)

			result, err := handler.processResponse(resp)

			if tt.expectErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if result != nil {
					t.Error("expected nil result when error occurs")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Error("expected non-nil result")
				}
				if len(result.Routes) == 0 {
					t.Error("expected routes in response")
				}
			}
		})
	}
}

func TestGeodistanceHandler_formatResponse(t *testing.T) {
	handler := &GeodistanceHandler{}

	tests := []struct {
		name         string
		responseBody *ResponseBody
		expectErr    bool
	}{
		{
			name: "valid response",
			responseBody: &ResponseBody{
				Routes: []Route{
					{
						DistanceMeters: 1000,
						Duration:       "5m",
						RouteLabels:    []string{"DEFAULT_ROUTE"},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "empty routes",
			responseBody: &ResponseBody{
				Routes: []Route{},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.formatResponse(tt.responseBody)

			if tt.expectErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if result != nil {
					t.Error("expected nil result when error occurs")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Error("expected non-nil result")
				}
				if len(result.Content) == 0 {
					t.Error("expected content in result")
				}
			}
		})
	}
}

func TestGeodistanceHandler_callDistanceMatrix(t *testing.T) {
	tests := []struct {
		name      string
		mockFunc  func(req *http.Request) (*http.Response, error)
		expectErr bool
	}{
		{
			name: "successful request",
			mockFunc: func(req *http.Request) (*http.Response, error) {
				return createMockResponse(http.StatusOK, createValidAPIResponse()), nil
			},
			expectErr: false,
		},
		{
			name: "HTTP error",
			mockFunc: func(req *http.Request) (*http.Response, error) {
				return nil, fmt.Errorf("network error")
			},
			expectErr: true,
		},
		{
			name: "API error response",
			mockFunc: func(req *http.Request) (*http.Response, error) {
				return createMockResponse(http.StatusBadRequest, `{"error": "Invalid request"}`), nil
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{DoFunc: tt.mockFunc}
			handler := &GeodistanceHandler{
				apiKey: "test-key",
				client: mockClient,
			}

			origins := []Origin{{Address: "New York"}}
			destinations := []Destination{{Address: "Los Angeles"}}

			result, err := handler.callDistanceMatrix(context.Background(), origins, destinations)

			if tt.expectErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if result != nil {
					t.Error("expected nil result when error occurs")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Error("expected non-nil result")
				}
			}
		})
	}
}

func TestGeodistanceHandler_handleDistanceCalculation(t *testing.T) {
	tests := []struct {
		name         string
		requestArgs  map[string]interface{}
		mockFunc     func(req *http.Request) (*http.Response, error)
		expectErr    bool
		expectedText string
	}{
		{
			name: "successful calculation",
			requestArgs: map[string]interface{}{
				"originAddress":      "New York",
				"destinationAddress": "Los Angeles",
			},
			mockFunc: func(req *http.Request) (*http.Response, error) {
				return createMockResponse(http.StatusOK, createValidAPIResponse()), nil
			},
			expectErr:    false,
			expectedText: "Route distance: 1000 meters, Duration: 5m",
		},
		{
			name: "missing origin address",
			requestArgs: map[string]interface{}{
				"destinationAddress": "Los Angeles",
			},
			mockFunc:  nil,
			expectErr: true,
		},
		{
			name: "missing destination address",
			requestArgs: map[string]interface{}{
				"originAddress": "New York",
			},
			mockFunc:  nil,
			expectErr: true,
		},
		{
			name: "empty origin address",
			requestArgs: map[string]interface{}{
				"originAddress":      "",
				"destinationAddress": "Los Angeles",
			},
			mockFunc:  nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockClient *MockHTTPClient
			if tt.mockFunc != nil {
				mockClient = &MockHTTPClient{DoFunc: tt.mockFunc}
			} else {
				mockClient = &MockHTTPClient{}
			}

			handler := &GeodistanceHandler{
				apiKey: "test-key",
				client: mockClient,
			}

			// Create mock request
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "calculate_distance",
					Arguments: tt.requestArgs,
				},
			}

			result, err := handler.handleDistanceCalculation(context.Background(), request)

			if tt.expectErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if result != nil {
					t.Error("expected nil result when error occurs")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Error("expected non-nil result")
				}
				if len(result.Content) == 0 {
					t.Error("expected content in result")
				}
				if textContent, ok := result.Content[0].(mcp.TextContent); ok {
					if textContent.Text != tt.expectedText {
						t.Errorf("expected text %q, got %q", tt.expectedText, textContent.Text)
					}
				} else {
					t.Error("expected text content")
				}
			}
		})
	}
}
