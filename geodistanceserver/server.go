package geodistanceserver

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var Version = "dev"

func GeodistanceServer() (*server.MCPServer, error) {
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
		mcp.WithDescription("Calculate distance between origin and destination addresses."),
		mcp.WithString("originAddress",
			mcp.Description("Address of origin"),
			mcp.Required(),
		),
		mcp.WithString("destinationAddress",
			mcp.Description("Address of destination"),
			mcp.Required(),
		),
	), h.handleDistanceCalculation)

	return s, nil
}
