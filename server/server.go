package server

import (
	"github.com/feenlace/mcp-1c/dump"
	"github.com/feenlace/mcp-1c/onec"
	"github.com/feenlace/mcp-1c/prompts"
	"github.com/feenlace/mcp-1c/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// New creates an MCP server with basic configuration and registers tools.
// If dumpSearcher is provided, the search_code tool will be registered.
func New(onecClient *onec.Client, dumpSearcher *dump.Searcher) *mcp.Server {
	s := mcp.NewServer(
		&mcp.Implementation{
			Name:    "mcp-1c",
			Version: "0.4.0-beta",
		},
		nil,
	)
	s.AddTool(tools.MetadataTool(), tools.NewMetadataHandler(onecClient))
	s.AddTool(tools.ObjectStructureTool(), tools.NewObjectStructureHandler(onecClient))
	s.AddTool(tools.QueryTool(), tools.NewQueryHandler(onecClient))
	if dumpSearcher != nil {
		s.AddTool(tools.SearchCodeTool(), tools.NewSearchCodeHandler(dumpSearcher))
	}
	s.AddTool(tools.FormStructureTool(), tools.NewFormStructureHandler(onecClient))
	s.AddTool(tools.ValidateQueryTool(), tools.NewValidateQueryHandler(onecClient))
	s.AddTool(tools.EventLogTool(), tools.NewEventLogHandler(onecClient))
	s.AddTool(tools.ConfigurationInfoTool(), tools.NewConfigurationInfoHandler(onecClient))
	tools.RegisterBSLHelp(s)
	prompts.RegisterAll(s)
	return s
}
