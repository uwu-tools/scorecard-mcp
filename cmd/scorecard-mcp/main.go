// Command scorecard-mcp runs an MCP server that exposes OpenSSF Scorecard
// security-posture data to MCP clients over stdio.
package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/uwu-tools/scorecard-mcp/internal/provider"
	"github.com/uwu-tools/scorecard-mcp/internal/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	p := provider.NewCachedREST("")
	s := server.New(p)

	if err := s.Run(ctx, &mcp.StdioTransport{}); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("scorecard-mcp: %v", err)
	}
}
