package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/kr/pretty"
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

func callDistanceMatrix(
	origins []Origin,
	destinations []Destination,
	travelMode string,
	routingPreference string,
) ([]MatrixElement, error) {
	// Load API key from environment variable
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GOOGLE_API_KEY environment variable not set")
	}

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

func main() {
	origins := []Origin{
		{
			Waypoint:       Waypoint{Location: Location{LatLng: LatLng{Latitude: 37.420761, Longitude: -122.081356}}},
			RouteModifiers: RouteModifiers{AvoidFerries: true},
		},
	}
	destinations := []Destination{
		{
			Waypoint: Waypoint{Location: Location{LatLng: LatLng{Latitude: 37.420999, Longitude: -122.086894}}},
		},
	}
	travelMode := "DRIVE"
	routingPreference := "TRAFFIC_AWARE"

	results, err := callDistanceMatrix(origins, destinations, travelMode, routingPreference)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	for _, elem := range results {
		fmt.Printf("Origin %d -> Destination %d: %d m, %s, %s\n",
			pretty.Formatter(elem.OriginIndex), pretty.Formatter(elem.DestinationIndex), pretty.Formatter(elem.DistanceMeters), pretty.Formatter(elem.Duration), pretty.Formatter(elem.Condition))
	}
}
