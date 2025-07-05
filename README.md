# geodistance
MCP server that calculates the distance between two geographic locations using the Google Maps API.

## Overview
This project provides both a standalone CLI application and an MCP (Model Context Protocol) server for calculating distances between geographic locations using Google's Routes API.

## Prerequisites
- Go 1.24.3 or higher
- Google Cloud Platform account with Maps API enabled
- Google Maps API key with Routes API access

## Setup

### 1. Install Dependencies
```bash
go mod download
```

### 2. Set Environment Variables
```bash
export GOOGLE_API_KEY=your_google_api_key_here
```

## Build

### Build Standalone CLI
```bash
go build -o geodistance .
```

### Build for Production (Optimized)
```bash
go build -ldflags="-s -w" -o geodistance .
```

### Build Docker Image
```bash
docker build -t geodistance .
```

## Test

### Run All Tests
```bash
go test ./...
```

### Run Tests with Verbose Output
```bash
go test -v ./...
```

### Run Tests with Coverage
```bash
go test -cover ./...
```

### Run Specific Package Tests
```bash
go test ./geodistanceserver/
```

## Run

### Run as MCP Server
```bash
# From source
go run .

# Using pre-built binary
./geodistance
```

### Run with Docker
```bash
docker run -e GOOGLE_API_KEY=your_api_key geodistance
```

## Usage

### MCP Server Mode
The server implements the MCP protocol and provides address-based distance calculations. Connect your MCP client to this server to calculate distances between two addresses.

### API Integration
- **Service**: Google Routes API v2
- **Authentication**: API key via `X-Goog-Api-Key` header
- **Input**: Address strings (automatically geocoded)
- **Output**: Distance in meters, duration, and route conditions

## Development

### Project Structure
```
geodistance/
├── main.go                    # CLI application entry point
├── geodistanceserver/         # MCP server implementation
│   ├── server.go             # Server setup and MCP handlers
│   ├── handler.go            # Distance calculation logic
│   ├── handler_test.go       # Handler unit tests
│   └── server_test.go        # Server integration tests
├── go.mod                    # Go module dependencies
├── go.sum                    # Dependency checksums
└── Dockerfile               # Container configuration
```

### Testing
The project includes comprehensive unit tests with mocking for external API calls. Tests cover both success and error scenarios.

### Contributing
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `go test ./...`
5. Build and test: `go build -o geodistance .`
6. Submit a pull request


## Sample Request Response Format
```bash
curl -X POST -d '{
  "origin":{
    "address": "Omaha, Nebraska"
  },
  "destination":{
    "address": "Lincoln, Nebraska"
  },
  "travelMode":"DRIVE",
  "routingPreference":"TRAFFIC_AWARE",
  "requestedReferenceRoutes": ["SHORTER_DISTANCE"],
  "languageCode": "en-US"
}' -H 'Content-Type: application/json' -H 'X-Goog-Api-Key: AIzaSyCEy8lyy1q1A0mLvPjhL0t0q7wNk2llc4o' -H 'X-Goog-FieldMask: routes.duration,routes.routeLabels,routes.distanceMeters' 'https://routes.googleapis.com/directions/v2:computeRoutes'
{
  "routes": [
    {
      "distanceMeters": 94475,
      "duration": "3288s",
      "routeLabels": [
        "DEFAULT_ROUTE"
      ]
    },
    {
      "distanceMeters": 87865,
      "duration": "4903s",
      "routeLabels": [
        "SHORTER_DISTANCE"
      ]
    }
  ]
}

```