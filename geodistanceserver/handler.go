package geodistanceserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
)

type LatLng struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Location struct {
	LatLng LatLng `json:"latLng"`
}

type Waypoint struct {
	Location Location `json:"location"`
}

type RouteModifiers struct {
	AvoidFerries bool `json:"avoid_ferries"`
}

type Origin struct {
	Waypoint       Waypoint       `json:"waypoint"`
	RouteModifiers RouteModifiers `json:"routeModifiers"`
}

type Destination struct {
	Waypoint Waypoint `json:"waypoint"`
}

type RequestBody struct {
	Origins           []Origin      `json:"origins"`
	Destinations      []Destination `json:"destinations"`
	TravelMode        string        `json:"travelMode"`
	RoutingPreference string        `json:"routingPreference"`
}

type MatrixElement struct {
	OriginIndex      int             `json:"originIndex"`
	DestinationIndex int             `json:"destinationIndex"`
	Status           json.RawMessage `json:"status"` // Use RawMessage if you don't care about details
	DistanceMeters   int             `json:"distanceMeters"`
	Duration         string          `json:"duration"` // format: "180s"
	Condition        string          `json:"condition"`
}

type GeodistanceHandler struct {
	apiKey string
}

func NewGeodistanceHandler() (*GeodistanceHandler, error) {
	// Load API key from environment variable
	googleApiKey := os.Getenv("GOOGLE_API_KEY")
	if googleApiKey == "" {
		return nil, fmt.Errorf("GOOGLE_API_KEY environment variable not set")
	}

	return &GeodistanceHandler{
		apiKey: googleApiKey,
	}, nil
}

func (gh *GeodistanceHandler) handleAddressDistanceCalculation(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	originAddress, err := request.RequireString("originAddress")
	if err != nil {
		return nil, err
	}
	destinationAddress, err := request.RequireString("destinationAddress")
	if err != nil {
		return nil, err
	}

	travelMode := "DRIVE"
	routingPreference := "TRAFFIC_AWARE"
}

func (gh *GeodistanceHandler) handleLatLongDistanceCalculation(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	originLatitude, err := request.RequireString("originLatitude")
	if err != nil {
		return nil, err
	}
	originLongitude, err := request.RequireString("originLongitude")
	if err != nil {
		return nil, err
	}

	destinationLatitude, err := request.RequireString("destinationLatitude")
	if err != nil {
		return nil, err
	}
	destinationLongitude, err := request.RequireString("destinationLongitude")
	if err != nil {
		return nil, err
	}

	origins := []Origin{
		{
			Waypoint:       Waypoint{Location: Location{LatLng: LatLng{Latitude: originLatitude, Longitude: originLongitude}}},
			RouteModifiers: RouteModifiers{AvoidFerries: true},
		},
	}
	destinations := []Destination{
		{
			Waypoint: Waypoint{Location: Location{LatLng: LatLng{Latitude: destinationLatitude, Longitude: destinationLongitude}}},
		},
	}
	travelMode := "DRIVE"
	routingPreference := "TRAFFIC_AWARE"
}

func callDistanceMatrix(
	origins []Origin,
	destinations []Destination,
	travelMode string,
	routingPreference string,
) ([]MatrixElement, error) {

	body := RequestBody{
		Origins:           origins,
		Destinations:      destinations,
		TravelMode:        travelMode,
		RoutingPreference: routingPreference,
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal json: %w", err)
	}

	url := "https://routes.googleapis.com/distanceMatrix/v2:computeRouteMatrix"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", apiKey)
	req.Header.Set("X-Goog-FieldMask", "originIndex,destinationIndex,duration,distanceMeters,status,condition")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var matrixElements []MatrixElement
	if err := json.Unmarshal(bodyBytes, &matrixElements); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return matrixElements, nil
}
