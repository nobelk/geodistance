package geodistanceserver

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var Version = "dev"

func NewGeodistanceServer() (*server.MCPServer, error) {
	h, err := NewGeodistanceHandler()
	if err != nil {
		return nil, err
	}

	s := server.NewMCPServer(
		"mcp-geodistance-server",
		Version,
		server.WithResourceCapabilities(true, true),
	)

	s.AddTool(mcp.NewTool(
		"calculate_distance",
		mcp.WithDescription("Calculate distance between origin and destination."),
		mcp.WithString("originLatitude",
			mcp.Description("Latitude of origin"),
			mcp.Required(),
		),
		mcp.WithString("originLongitude",
			mcp.Description("Longitude of Origin"),
			mcp.Required(),
		),
		mcp.WithString("destinationLatitude",
			mcp.Description("Latitude of destination"),
			mcp.Required(),
		),
		mcp.WithString("destinationLongitude",
			mcp.Description("Longitude of Destination"),
			mcp.Required(),
		),
	), h.handleDistanceCalculation)

	return s, nil
}
