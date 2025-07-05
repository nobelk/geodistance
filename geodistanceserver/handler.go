package geodistanceserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

type Origin struct {
	Address string `json:"address"`
}

type Destination struct {
	Address string `json:"address"`
}

type RequestBody struct {
	Origins                  []Origin      `json:"origins"`
	Destinations             []Destination `json:"destinations"`
	TravelMode               string        `json:"travelMode"`
	RoutingPreference        string        `json:"routingPreference"`
	RequestedReferenceRoutes []string      `json:"requestedReferenceRoutes"`
	LanguageCode             string        `json:"languageCode"`
}

type ResponseBody struct {
	Routes []Route `json:"routes"`
}

type Route struct {
	DistanceMeters int      `json:"distanceMeters"`
	Duration       string   `json:"duration"`
	RouteLabels    []string `json:"routeLabels"`
}

// HTTPClient interface for testability
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type GeodistanceHandler struct {
	apiKey string
	client HTTPClient
}

func NewGeodistanceHandler() (*GeodistanceHandler, error) {
	return NewGeodistanceHandlerWithClient(&http.Client{
		Timeout: 30 * time.Second,
	})
}

func NewGeodistanceHandlerWithClient(client HTTPClient) (*GeodistanceHandler, error) {
	// Load API key from environment variable
	googleApiKey := os.Getenv("GOOGLE_API_KEY")
	if googleApiKey == "" {
		return nil, fmt.Errorf("GOOGLE_API_KEY environment variable not set")
	}

	return &GeodistanceHandler{
		apiKey: googleApiKey,
		client: client,
	}, nil
}

func (gh *GeodistanceHandler) handleDistanceCalculation(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	originAddress, err := request.RequireString("originAddress")
	if err != nil {
		return nil, fmt.Errorf("missing origin address: %w", err)
	}

	destinationAddress, err := request.RequireString("destinationAddress")
	if err != nil {
		return nil, fmt.Errorf("missing destination address: %w", err)
	}

	if err := gh.validateAddresses(originAddress, destinationAddress); err != nil {
		return nil, err
	}

	origins := []Origin{{Address: originAddress}}
	destinations := []Destination{{Address: destinationAddress}}

	responseBody, err := gh.callDistanceMatrix(ctx, origins, destinations)
	if err != nil {
		return nil, err
	}

	return gh.formatResponse(responseBody)
}

func (gh *GeodistanceHandler) validateAddresses(origin, destination string) error {
	if origin == "" {
		return fmt.Errorf("origin address cannot be empty")
	}
	if destination == "" {
		return fmt.Errorf("destination address cannot be empty")
	}
	return nil
}

func (gh *GeodistanceHandler) buildRequestBody(origins []Origin, destinations []Destination) *RequestBody {
	return &RequestBody{
		Origins:                  origins,
		Destinations:             destinations,
		TravelMode:               "DRIVE",
		RoutingPreference:        "TRAFFIC_AWARE",
		RequestedReferenceRoutes: []string{"SHORTER_DISTANCE"},
		LanguageCode:             "en-US",
	}
}

func (gh *GeodistanceHandler) createRequest(ctx context.Context, body *RequestBody) (*http.Request, error) {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal json: %w", err)
	}

	url := "https://routes.googleapis.com/distanceMatrix/v2:computeRouteMatrix"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", gh.apiKey)
	req.Header.Set("X-Goog-FieldMask", "routes.duration,routes.routeLabels,routes.distanceMeters")

	return req, nil
}

func (gh *GeodistanceHandler) processResponse(resp *http.Response) (*ResponseBody, error) {
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var responseBody ResponseBody
	if err := json.Unmarshal(bodyBytes, &responseBody); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(responseBody.Routes) == 0 {
		return nil, fmt.Errorf("no routes found in response")
	}

	return &responseBody, nil
}

func (gh *GeodistanceHandler) formatResponse(responseBody *ResponseBody) (*mcp.CallToolResult, error) {
	if len(responseBody.Routes) == 0 {
		return nil, fmt.Errorf("no routes available")
	}

	route := responseBody.Routes[0]
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Route distance: %d meters, Duration: %s", route.DistanceMeters, route.Duration),
			},
		},
	}, nil
}

func (gh *GeodistanceHandler) callDistanceMatrix(
	ctx context.Context,
	origins []Origin,
	destinations []Destination,
) (*ResponseBody, error) {
	body := gh.buildRequestBody(origins, destinations)

	req, err := gh.createRequest(ctx, body)
	if err != nil {
		return nil, err
	}

	resp, err := gh.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	return gh.processResponse(resp)
}
