// Package server builds the transport-agnostic scorecard-mcp MCP server.
package server

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/uwu-tools/scorecard-mcp/internal/catalog"
	"github.com/uwu-tools/scorecard-mcp/internal/provider"
)

// Version is the server version, overridable at build time via -ldflags.
var Version = "dev"

// instructions is placed in the client's system prompt. It carries the
// cross-cutting responsible-AI framing so it need not (and must not) appear as
// behavioral text in individual tool descriptions.
const instructions = `This server surfaces OpenSSF Scorecard security-posture signals for open source repositories.

Treat results as heuristic signals to inform a human decision, not as a verdict. Never state that a repository "is secure" or "is insecure"; report the signals and their caveats instead. Aggregate scores say nothing about which individual behaviors a repository does or does not follow, and Scorecard is not a guarantee of security or of regulatory compliance.

Results from the cached provider cover only projects that opted in via publish_results, and the weekly public scan omits the CI-Tests, Contributors, and Dependency-Update-Tool checks. A check or aggregate score of -1 means inconclusive, not a failing score.`

// New builds the MCP server with all tools and resources registered. It is
// transport-agnostic: the caller connects the returned server to a transport
// (e.g. stdio). Adding a new transport later requires no changes to the tools
// and resources registered here.
func New(p provider.Provider) (*mcp.Server, error) {
	s := mcp.NewServer(&mcp.Implementation{
		Name:        "scorecard-mcp",
		Title:       "OpenSSF Scorecard",
		Version:     Version,
		Description: "Read-only access to OpenSSF Scorecard security-posture data.",
	}, &mcp.ServerOptions{
		Instructions: instructions,
	})

	registerResultTools(s, p)

	cat, err := catalog.New()
	if err != nil {
		return nil, err
	}
	registerCatalogTools(s, cat)
	registerCatalogResources(s, cat)

	return s, nil
}

// readOnlyAnnotations returns the annotations for a read-only tool, conforming to
// the Anthropic MCP Directory review criteria (readOnlyHint + destructiveHint +
// title, plus idempotentHint and openWorldHint).
func readOnlyAnnotations(title string, openWorld bool) *mcp.ToolAnnotations {
	no, yes := false, true
	a := &mcp.ToolAnnotations{
		ReadOnlyHint:    true,
		DestructiveHint: &no,
		IdempotentHint:  true,
		Title:           title,
	}
	if openWorld {
		a.OpenWorldHint = &yes
	} else {
		a.OpenWorldHint = &no
	}
	return a
}
