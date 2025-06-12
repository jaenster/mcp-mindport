package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"mcp-mindport/internal/config"
	"mcp-mindport/internal/domain"
	"mcp-mindport/internal/search"
	"mcp-mindport/internal/storage"

	"golang.org/x/crypto/blake2b"
)

type Server struct {
	storage       *storage.BadgerStore
	searchEngine  *search.BleveSearch
	cliTools      *search.CLISearchTools
	domainManager *domain.DomainManager
	config        *config.Config
}

type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ClientInfo      ClientInfo             `json:"clientInfo"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitializeResult struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ServerInfo      ServerInfo             `json:"serverInfo"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func NewServer(storage *storage.BadgerStore, searchEngine *search.BleveSearch, config *config.Config) *Server {
	cliTools := search.NewCLISearchTools(storage, searchEngine)
	
	// Initialize domain manager
	domainConfig := &domain.DomainConfig{
		DefaultDomain:    config.Domain.DefaultDomain,
		IsolationMode:    config.Domain.IsolationMode,
		AllowCrossDomain: config.Domain.AllowCrossDomain,
	}
	domainManager := domain.NewDomainManager(domainConfig)
	
	return &Server{
		storage:       storage,
		searchEngine:  searchEngine,
		cliTools:      cliTools,
		domainManager: domainManager,
		config:        config,
	}
}

func (s *Server) Start(ctx context.Context) error {
	// Read from stdin and write to stdout (MCP stdio transport)
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	log.Printf("MCP server starting stdio transport")

	for {
		select {
		case <-ctx.Done():
			log.Printf("MCP server context cancelled")
			return ctx.Err()
		default:
			var request MCPRequest
			if err := decoder.Decode(&request); err != nil {
				if err == io.EOF {
					log.Printf("MCP server received EOF, shutting down")
					return nil
				}
				log.Printf("Failed to decode request: %v", err)
				continue
			}

			log.Printf("MCP server received request: method=%s id=%v", request.Method, request.ID)

			response := s.HandleRequest(ctx, &request)
			
			// Don't send response for notifications (response will be nil)
			if response != nil {
				log.Printf("MCP server sending response: id=%v error=%v", response.ID, response.Error != nil)
				
				if err := encoder.Encode(response); err != nil {
					log.Printf("Failed to encode response: %v", err)
					// Don't continue on encoding errors - this might indicate client disconnect
					return fmt.Errorf("response encoding failed: %v", err)
				}
				
				// Flush output to ensure response is sent immediately
				if err := os.Stdout.Sync(); err != nil {
					log.Printf("Failed to flush stdout: %v", err)
				}
			} else {
				log.Printf("No response sent for notification/invalid request")
			}
		}
	}
}

// getValidID ensures we have a proper ID for JSON-RPC responses
func getValidID(requestID interface{}) interface{} {
	if requestID == nil {
		return 0 // Default for null IDs
	}
	return requestID
}

func (s *Server) HandleRequest(ctx context.Context, request *MCPRequest) *MCPResponse {
	switch request.Method {
	case "initialize":
		return s.handleInitialize(request)
	case "notifications/initialized":
		// This is a notification - no response should be sent
		log.Printf("Received initialized notification")
		return nil
	case "resources/list":
		return s.handleResourcesList(ctx, request)
	case "resources/read":
		return s.handleResourcesRead(ctx, request)
	case "tools/list":
		return s.handleToolsList(request)
	case "tools/call":
		return s.handleToolsCall(ctx, request)
	case "prompts/list":
		return s.handlePromptsList(ctx, request)
	case "prompts/get":
		return s.handlePromptsGet(ctx, request)
	default:
		// Only send error response if this is not a notification (has an ID)
		if request.ID == nil {
			log.Printf("Ignoring unknown notification: %s", request.Method)
			return nil
		}
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32601,
				Message: "Method not found",
			},
		}
	}
}

func (s *Server) handleInitialize(request *MCPRequest) *MCPResponse {
	// Enhanced initialize handling for Claude Code compatibility
	log.Printf("Handling initialize request from client")
	
	// Extract client info for debugging
	if params, ok := request.Params.(map[string]interface{}); ok {
		if clientInfo, ok := params["clientInfo"].(map[string]interface{}); ok {
			if name, ok := clientInfo["name"].(string); ok {
				log.Printf("Client name: %s", name)
			}
		}
		if version, ok := params["protocolVersion"].(string); ok {
			log.Printf("Client protocol version: %s", version)
		}
	}

	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: map[string]interface{}{
			"resources": map[string]interface{}{
				"subscribe":   true,
				"listChanged": true,
			},
			"tools": map[string]interface{}{},
			"prompts": map[string]interface{}{},
		},
		ServerInfo: ServerInfo{
			Name:    "mcp-mindport",
			Version: "1.0.0",
		},
	}

	response := &MCPResponse{
		JSONRPC: "2.0",
		ID:      getValidID(request.ID),
		Result:  result,
	}

	log.Printf("Initialize response prepared successfully")
	return response
}

func (s *Server) handleResourcesList(ctx context.Context, request *MCPRequest) *MCPResponse {
	resources, err := s.storage.ListResources(ctx, 100, 0)
	if err != nil {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32603,
				Message: "Internal error",
				Data:    err.Error(),
			},
		}
	}

	var mcpResources []map[string]interface{}
	for _, resource := range resources {
		mcpResources = append(mcpResources, map[string]interface{}{
			"uri":         fmt.Sprintf("mindport://resource/%s", resource.ID),
			"name":        resource.Title,
			"description": fmt.Sprintf("Resource: %s", resource.Title),
			"mimeType":    "text/plain",
		})
	}

	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      getValidID(request.ID),
		Result: map[string]interface{}{
			"resources": mcpResources,
		},
	}
}

func (s *Server) handleResourcesRead(ctx context.Context, request *MCPRequest) *MCPResponse {
	params, ok := request.Params.(map[string]interface{})
	if !ok {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid params",
			},
		}
	}

	uri, ok := params["uri"].(string)
	if !ok {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32602,
				Message: "URI is required",
			},
		}
	}

	// Extract resource ID from URI (mindport://resource/{id})
	parts := strings.Split(uri, "/")
	if len(parts) < 4 || parts[0] != "mindport:" || parts[2] != "resource" {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid URI format",
			},
		}
	}

	resourceID := parts[3]
	if resourceID == "" {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid URI format",
			},
		}
	}
	
	// Use domain-aware resource retrieval
	currentDomain := s.domainManager.GetCurrentDomain()
	resource, err := s.storage.GetResourceInDomain(ctx, resourceID, currentDomain)
	if err != nil {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32603,
				Message: "Resource not found",
				Data:    err.Error(),
			},
		}
	}

	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      getValidID(request.ID),
		Result: map[string]interface{}{
			"contents": []map[string]interface{}{
				{
					"uri":      uri,
					"mimeType": "text/plain",
					"text":     resource.Content,
				},
			},
		},
	}
}

func (s *Server) handleToolsList(request *MCPRequest) *MCPResponse {
	tools := []map[string]interface{}{
		{
			"name":        "store_resource",
			"description": "Store a new resource with content and metadata",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"title": map[string]interface{}{
						"type":        "string",
						"description": "Title of the resource",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Content of the resource",
					},
					"type": map[string]interface{}{
						"type":        "string",
						"description": "Type of the resource",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Tags for the resource",
					},
					"metadata": map[string]interface{}{
						"type":        "object",
						"description": "Additional metadata",
					},
				},
				"required": []string{"title", "content"},
			},
		},
		{
			"name":        "search_resources",
			"description": "Search for resources using optimized token-efficient search",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of results",
						"default":     10,
					},
					"type": map[string]interface{}{
						"type":        "string",
						"description": "Filter by type (resource or prompt)",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Filter by tags",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			"name":        "store_prompt",
			"description": "Store a new prompt template",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the prompt",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Description of the prompt",
					},
					"template": map[string]interface{}{
						"type":        "string",
						"description": "Prompt template content",
					},
					"variables": map[string]interface{}{
						"type":        "object",
						"description": "Template variables",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Tags for the prompt",
					},
				},
				"required": []string{"name", "template"},
			},
		},
		{
			"name":        "advanced_search",
			"description": "Advanced search with IDE-like features: fuzzy, regex, semantic search",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query",
					},
					"mode": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"exact", "fuzzy", "regex", "wildcard", "semantic", "smart"},
						"description": "Search mode: exact, fuzzy, regex, wildcard, semantic, or smart (auto-detect)",
						"default":     "smart",
					},
					"case_sensitive": map[string]interface{}{
						"type":        "boolean",
						"description": "Case sensitive search",
					},
					"whole_words": map[string]interface{}{
						"type":        "boolean",
						"description": "Match whole words only",
					},
					"type": map[string]interface{}{
						"type":        "string",
						"description": "Filter by type (resource/prompt)",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Filter by tags",
					},
					"fields": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Search in specific fields (title, content, tags)",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum results",
						"default":     20,
					},
					"sort_by": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"relevance", "date", "title", "type"},
						"description": "Sort results by",
					},
					"highlight": map[string]interface{}{
						"type":        "boolean",
						"description": "Highlight matches",
					},
					"domains": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Filter by domains",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			"name":        "grep",
			"description": "grep-like search with familiar CLI options",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"pattern": map[string]interface{}{
						"type":        "string",
						"description": "Search pattern",
					},
					"ignore_case": map[string]interface{}{
						"type":        "boolean",
						"description": "Ignore case (like grep -i)",
					},
					"invert_match": map[string]interface{}{
						"type":        "boolean",
						"description": "Invert match (like grep -v)",
					},
					"line_numbers": map[string]interface{}{
						"type":        "boolean",
						"description": "Show line numbers (like grep -n)",
					},
					"count": map[string]interface{}{
						"type":        "boolean",
						"description": "Show only count (like grep -c)",
					},
					"context": map[string]interface{}{
						"type":        "integer",
						"description": "Lines of context (like grep -C)",
					},
					"whole_words": map[string]interface{}{
						"type":        "boolean",
						"description": "Match whole words (like grep -w)",
					},
					"extended": map[string]interface{}{
						"type":        "boolean",
						"description": "Extended regex (like grep -E)",
					},
					"fixed": map[string]interface{}{
						"type":        "boolean",
						"description": "Fixed strings (like grep -F)",
					},
					"max_matches": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum matches (like grep -m)",
						"default":     1000,
					},
					"domains": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Filter by domains",
					},
				},
				"required": []string{"pattern"},
			},
		},
		{
			"name":        "find",
			"description": "find-like search for resources by metadata",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name pattern (like find -name)",
					},
					"type": map[string]interface{}{
						"type":        "string",
						"description": "Type filter: f=resource, d=prompt (like find -type)",
					},
					"size": map[string]interface{}{
						"type":        "string",
						"description": "Size filter: +100k, -1M (like find -size)",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Filter by tags",
					},
					"content_type": map[string]interface{}{
						"type":        "string",
						"description": "Filter by content type",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum results",
						"default":     1000,
					},
					"domains": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Filter by domains",
					},
				},
			},
		},
		{
			"name":        "ripgrep",
			"description": "ripgrep-style advanced search with smart defaults",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"pattern": map[string]interface{}{
						"type":        "string",
						"description": "Search pattern",
					},
					"ignore_case": map[string]interface{}{
						"type":        "boolean",
						"description": "Ignore case (like rg -i)",
					},
					"smart_case": map[string]interface{}{
						"type":        "boolean",
						"description": "Smart case matching (like rg -S)",
					},
					"word_regexp": map[string]interface{}{
						"type":        "boolean",
						"description": "Word boundaries (like rg -w)",
					},
					"fixed": map[string]interface{}{
						"type":        "boolean",
						"description": "Fixed strings (like rg -F)",
					},
					"count": map[string]interface{}{
						"type":        "boolean",
						"description": "Show only count (like rg -c)",
					},
					"files_with_matches": map[string]interface{}{
						"type":        "boolean",
						"description": "Show only matching files (like rg -l)",
					},
					"line_number": map[string]interface{}{
						"type":        "boolean",
						"description": "Show line numbers (like rg -n)",
					},
					"context": map[string]interface{}{
						"type":        "integer",
						"description": "Lines of context (like rg -C)",
					},
					"max_count": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum matches per resource (like rg -m)",
					},
					"type": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Filter by type (like rg -t)",
					},
					"domains": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Filter by domains",
					},
				},
				"required": []string{"pattern"},
			},
		},
		{
			"name":        "create_domain",
			"description": "Create a new domain/project for organizing resources",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Domain ID (lowercase, alphanumeric, hyphens)",
						"pattern":     "^[a-z0-9][a-z0-9\\-_]*[a-z0-9]$",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Human-readable domain name",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Domain description",
					},
					"parent": map[string]interface{}{
						"type":        "string",
						"description": "Parent domain ID for hierarchical organization",
					},
				},
				"required": []string{"id", "name"},
			},
		},
		{
			"name":        "list_domains",
			"description": "List all available domains/projects",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"parent": map[string]interface{}{
						"type":        "string",
						"description": "Filter by parent domain",
					},
					"include_stats": map[string]interface{}{
						"type":        "boolean",
						"description": "Include resource/prompt counts",
					},
				},
			},
		},
		{
			"name":        "switch_domain",
			"description": "Switch to a different domain context",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"domain": map[string]interface{}{
						"type":        "string",
						"description": "Domain ID to switch to",
					},
				},
				"required": []string{"domain"},
			},
		},
		{
			"name":        "domain_stats",
			"description": "Get statistics for a domain",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"domain": map[string]interface{}{
						"type":        "string",
						"description": "Domain ID (defaults to current domain)",
					},
				},
			},
		},
		{
			"name":        "get_resource",
			"description": "Get a resource by ID with shorthand domain notation support",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Resource ID with optional domain notation (::abc123 for default, proj1:abc123 for specific domain, or abc123 for default)",
					},
				},
				"required": []string{"id"},
			},
		},
		{
			"name":        "get_prompt",
			"description": "Get a prompt by ID with shorthand domain notation support",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Prompt ID with optional domain notation (::abc123 for default, proj1:abc123 for specific domain, or abc123 for default)",
					},
				},
				"required": []string{"id"},
			},
		},
	}

	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      getValidID(request.ID),
		Result: map[string]interface{}{
			"tools": tools,
		},
	}
}

func (s *Server) handleToolsCall(ctx context.Context, request *MCPRequest) *MCPResponse {
	params, ok := request.Params.(map[string]interface{})
	if !ok {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid params",
			},
		}
	}

	name, ok := params["name"].(string)
	if !ok {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32602,
				Message: "Tool name is required",
			},
		}
	}

	arguments, ok := params["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}

	switch name {
	case "store_resource":
		return s.handleStoreResource(ctx, request, arguments)
	case "search_resources":
		return s.handleSearchResources(ctx, request, arguments)
	case "store_prompt":
		return s.handleStorePrompt(ctx, request, arguments)
	case "advanced_search":
		return s.handleAdvancedSearch(ctx, request, arguments)
	case "grep":
		return s.handleGrep(ctx, request, arguments)
	case "find":
		return s.handleFind(ctx, request, arguments)
	case "ripgrep":
		return s.handleRipgrep(ctx, request, arguments)
	case "create_domain":
		return s.handleCreateDomain(ctx, request, arguments)
	case "list_domains":
		return s.handleListDomains(ctx, request, arguments)
	case "switch_domain":
		return s.handleSwitchDomain(ctx, request, arguments)
	case "domain_stats":
		return s.handleDomainStats(ctx, request, arguments)
	case "get_resource":
		return s.handleGetResource(ctx, request, arguments)
	case "get_prompt":
		return s.handleGetPrompt(ctx, request, arguments)
	default:
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32601,
				Message: "Tool not found",
			},
		}
	}
}

func (s *Server) handleStoreResource(ctx context.Context, request *MCPRequest, args map[string]interface{}) *MCPResponse {
	title, ok := args["title"].(string)
	if !ok {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32602,
				Message: "Title is required",
			},
		}
	}

	content, ok := args["content"].(string)
	if !ok {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32602,
				Message: "Content is required",
			},
		}
	}

	// Generate ID based on content hash
	hash := blake2b.Sum256([]byte(title + content))
	id := fmt.Sprintf("%x", hash[:8])

	resource := &storage.Resource{
		ID:      id,
		Title:   title,
		Content: content,
		Domain:  s.domainManager.GetCurrentDomain(),
	}

	if resourceType, ok := args["type"].(string); ok {
		resource.Type = resourceType
	}

	if tags, ok := args["tags"].([]interface{}); ok {
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				resource.Tags = append(resource.Tags, tagStr)
			}
		}
	}

	if metadata, ok := args["metadata"].(map[string]interface{}); ok {
		resource.Metadata = metadata
	}

	// Store resource
	if err := s.storage.StoreResourceInDomain(ctx, resource, resource.Domain); err != nil {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32603,
				Message: "Failed to store resource",
				Data:    err.Error(),
			},
		}
	}

	// Index for search
	if err := s.searchEngine.IndexResource(ctx, resource); err != nil {
		log.Printf("Failed to index resource: %v", err)
	}

	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      getValidID(request.ID),
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": fmt.Sprintf("Resource stored successfully with ID: %s", id),
				},
			},
		},
	}
}

func (s *Server) handleSearchResources(ctx context.Context, request *MCPRequest, args map[string]interface{}) *MCPResponse {
	query, ok := args["query"].(string)
	if !ok {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32602,
				Message: "Query is required",
			},
		}
	}

	limit := 10
	if limitFloat, ok := args["limit"].(float64); ok {
		limit = int(limitFloat)
	}

	// Use optimized search for token efficiency
	result, err := s.searchEngine.OptimizedSearch(ctx, query, limit)
	if err != nil {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32603,
				Message: "Search failed",
				Data:    err.Error(),
			},
		}
	}

	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      getValidID(request.ID),
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": result,
				},
			},
		},
	}
}

func (s *Server) handleStorePrompt(ctx context.Context, request *MCPRequest, args map[string]interface{}) *MCPResponse {
	name, ok := args["name"].(string)
	if !ok {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32602,
				Message: "Name is required",
			},
		}
	}

	template, ok := args["template"].(string)
	if !ok {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32602,
				Message: "Template is required",
			},
		}
	}

	// Generate ID based on name hash
	hash := blake2b.Sum256([]byte(name + template))
	id := fmt.Sprintf("%x", hash[:8])

	prompt := &storage.Prompt{
		ID:       id,
		Name:     name,
		Template: template,
		Domain:   s.domainManager.GetCurrentDomain(),
	}

	if description, ok := args["description"].(string); ok {
		prompt.Description = description
	}

	if variables, ok := args["variables"].(map[string]interface{}); ok {
		prompt.Variables = make(map[string]string)
		for k, v := range variables {
			if vStr, ok := v.(string); ok {
				prompt.Variables[k] = vStr
			}
		}
	}

	if tags, ok := args["tags"].([]interface{}); ok {
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				prompt.Tags = append(prompt.Tags, tagStr)
			}
		}
	}

	// Store prompt
	if err := s.storage.StorePromptInDomain(ctx, prompt, prompt.Domain); err != nil {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32603,
				Message: "Failed to store prompt",
				Data:    err.Error(),
			},
		}
	}

	// Index for search
	if err := s.searchEngine.IndexPrompt(ctx, prompt); err != nil {
		log.Printf("Failed to index prompt: %v", err)
	}

	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      getValidID(request.ID),
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": fmt.Sprintf("Prompt stored successfully with ID: %s", id),
				},
			},
		},
	}
}

func (s *Server) handlePromptsList(ctx context.Context, request *MCPRequest) *MCPResponse {
	prompts, err := s.storage.ListPrompts(ctx, 100, 0)
	if err != nil {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32603,
				Message: "Internal error",
				Data:    err.Error(),
			},
		}
	}

	var mcpPrompts []map[string]interface{}
	for _, prompt := range prompts {
		arguments := make([]map[string]interface{}, 0)
		for varName, varDesc := range prompt.Variables {
			arguments = append(arguments, map[string]interface{}{
				"name":        varName,
				"description": varDesc,
				"required":    true,
			})
		}

		mcpPrompts = append(mcpPrompts, map[string]interface{}{
			"name":        prompt.Name,
			"description": prompt.Description,
			"arguments":   arguments,
		})
	}

	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      getValidID(request.ID),
		Result: map[string]interface{}{
			"prompts": mcpPrompts,
		},
	}
}

func (s *Server) handlePromptsGet(ctx context.Context, request *MCPRequest) *MCPResponse {
	params, ok := request.Params.(map[string]interface{})
	if !ok {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid params",
			},
		}
	}

	name, ok := params["name"].(string)
	if !ok {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32602,
				Message: "Name is required",
			},
		}
	}

	// Find prompt by name (we'll need to search through all prompts)
	prompts, err := s.storage.ListPrompts(ctx, 1000, 0)
	if err != nil {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32603,
				Message: "Internal error",
				Data:    err.Error(),
			},
		}
	}

	var foundPrompt *storage.Prompt
	for _, prompt := range prompts {
		if prompt.Name == name {
			foundPrompt = prompt
			break
		}
	}

	if foundPrompt == nil {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32603,
				Message: "Prompt not found",
			},
		}
	}

	// Process template with provided arguments
	template := foundPrompt.Template
	if arguments, ok := params["arguments"].(map[string]interface{}); ok {
		for key, value := range arguments {
			if valueStr, ok := value.(string); ok {
				template = strings.ReplaceAll(template, fmt.Sprintf("{{%s}}", key), valueStr)
			}
		}
	}

	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      getValidID(request.ID),
		Result: map[string]interface{}{
			"description": foundPrompt.Description,
			"messages": []map[string]interface{}{
				{
					"role": "user",
					"content": map[string]interface{}{
						"type": "text",
						"text": template,
					},
				},
			},
		},
	}
}

// Handler for advanced search
func (s *Server) handleAdvancedSearch(ctx context.Context, request *MCPRequest, args map[string]interface{}) *MCPResponse {
	query, ok := args["query"].(string)
	if !ok {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32602,
				Message: "Query is required",
			},
		}
	}

	// Build advanced search query
	searchQuery := &search.AdvancedSearchQuery{
		Query: query,
		Limit: 20,
	}

	// Apply optional parameters
	if mode, ok := args["mode"].(string); ok {
		searchQuery.Mode = mode
	}
	if caseSensitive, ok := args["case_sensitive"].(bool); ok {
		searchQuery.CaseSensitive = caseSensitive
	}
	if wholeWords, ok := args["whole_words"].(bool); ok {
		searchQuery.WholeWords = wholeWords
	}
	if searchType, ok := args["type"].(string); ok {
		searchQuery.Type = searchType
	}
	if limit, ok := args["limit"].(float64); ok {
		searchQuery.Limit = int(limit)
	}
	if sortBy, ok := args["sort_by"].(string); ok {
		searchQuery.SortBy = sortBy
	}
	if highlight, ok := args["highlight"].(bool); ok {
		searchQuery.Highlight = highlight
	}

	// Handle tags
	if tagsInterface, ok := args["tags"].([]interface{}); ok {
		for _, tag := range tagsInterface {
			if tagStr, ok := tag.(string); ok {
				searchQuery.Tags = append(searchQuery.Tags, tagStr)
			}
		}
	}

	// Handle fields
	if fieldsInterface, ok := args["fields"].([]interface{}); ok {
		for _, field := range fieldsInterface {
			if fieldStr, ok := field.(string); ok {
				searchQuery.Fields = append(searchQuery.Fields, fieldStr)
			}
		}
	}
	
	// Handle domains
	if domainsInterface, ok := args["domains"].([]interface{}); ok {
		for _, domain := range domainsInterface {
			if domainStr, ok := domain.(string); ok {
				searchQuery.Domains = append(searchQuery.Domains, domainStr)
			}
		}
	}

	// Perform search
	results, stats, err := s.searchEngine.AdvancedSearch(ctx, searchQuery)
	if err != nil {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32603,
				Message: "Advanced search failed",
				Data:    err.Error(),
			},
		}
	}

	// Format results for token efficiency
	var output strings.Builder
	output.WriteString(fmt.Sprintf("🔍 Found %d results in %v\n\n", stats.TotalResults, stats.SearchTime))

	for i, result := range results {
		output.WriteString(fmt.Sprintf("%d. [%s] %s (Score: %.2f)\n", 
			i+1, result.Type, result.Title, result.Score))
		
		if result.Snippet != "" {
			output.WriteString(fmt.Sprintf("   %s\n", result.Snippet))
		}
		
		if len(result.Highlights) > 0 {
			output.WriteString(fmt.Sprintf("   Highlights: %s\n", strings.Join(result.Highlights, ", ")))
		}
		
		if len(result.Tags) > 0 {
			output.WriteString(fmt.Sprintf("   Tags: %s\n", strings.Join(result.Tags, ", ")))
		}
		
		if i < len(results)-1 {
			output.WriteString("\n")
		}
	}

	// Add statistics
	if len(stats.TypeBreakdown) > 0 {
		output.WriteString("\n📊 Type breakdown: ")
		for typ, count := range stats.TypeBreakdown {
			output.WriteString(fmt.Sprintf("%s:%d ", typ, count))
		}
	}

	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      getValidID(request.ID),
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": output.String(),
				},
			},
		},
	}
}

// Handler for grep-style search
func (s *Server) handleGrep(ctx context.Context, request *MCPRequest, args map[string]interface{}) *MCPResponse {
	pattern, ok := args["pattern"].(string)
	if !ok {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32602,
				Message: "Pattern is required",
			},
		}
	}

	// Build grep options
	opts := &search.GrepOptions{
		Pattern:    pattern,
		MaxMatches: 1000,
	}

	// Apply optional parameters
	if ignoreCase, ok := args["ignore_case"].(bool); ok {
		opts.IgnoreCase = ignoreCase
	}
	if invertMatch, ok := args["invert_match"].(bool); ok {
		opts.InvertMatch = invertMatch
	}
	if lineNumbers, ok := args["line_numbers"].(bool); ok {
		opts.LineNumbers = lineNumbers
	}
	if count, ok := args["count"].(bool); ok {
		opts.Count = count
	}
	if context, ok := args["context"].(float64); ok {
		opts.Context = int(context)
	}
	if wholeWords, ok := args["whole_words"].(bool); ok {
		opts.WholeWords = wholeWords
	}
	if extended, ok := args["extended"].(bool); ok {
		opts.Extended = extended
	}
	if fixed, ok := args["fixed"].(bool); ok {
		opts.Fixed = fixed
	}
	if maxMatches, ok := args["max_matches"].(float64); ok {
		opts.MaxMatches = int(maxMatches)
	}
	
	// Handle domain filtering
	if domainsInterface, ok := args["domains"].([]interface{}); ok {
		for _, domain := range domainsInterface {
			if domainStr, ok := domain.(string); ok {
				opts.Domains = append(opts.Domains, domainStr)
			}
		}
	}

	// Perform grep search
	results, err := s.cliTools.Grep(ctx, opts)
	if err != nil {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32603,
				Message: "Grep search failed",
				Data:    err.Error(),
			},
		}
	}

	// Format results in grep style
	var output strings.Builder
	
	if opts.Count {
		output.WriteString(fmt.Sprintf("Total matches: %d\n", results[0].MatchCount))
	} else {
		for _, result := range results {
			if opts.LineNumbers && result.LineNumber > 0 {
				output.WriteString(fmt.Sprintf("%s:%d:%s\n", result.Title, result.LineNumber, result.MatchedLine))
			} else {
				output.WriteString(fmt.Sprintf("%s:%s\n", result.Title, result.MatchedLine))
			}
			
			// Add context if available
			for _, contextLine := range result.Context {
				output.WriteString(fmt.Sprintf("%s\n", contextLine))
			}
		}
	}

	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      getValidID(request.ID),
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": output.String(),
				},
			},
		},
	}
}

// Handler for find-style search
func (s *Server) handleFind(ctx context.Context, request *MCPRequest, args map[string]interface{}) *MCPResponse {
	opts := &search.FindOptions{
		Limit: 1000,
	}

	// Apply parameters
	if name, ok := args["name"].(string); ok {
		opts.Name = name
	}
	if findType, ok := args["type"].(string); ok {
		opts.Type = findType
	}
	if size, ok := args["size"].(string); ok {
		opts.Size = size
	}
	if contentType, ok := args["content_type"].(string); ok {
		opts.ContentType = contentType
	}
	if limit, ok := args["limit"].(float64); ok {
		opts.Limit = int(limit)
	}

	// Handle tags
	if tagsInterface, ok := args["tags"].([]interface{}); ok {
		for _, tag := range tagsInterface {
			if tagStr, ok := tag.(string); ok {
				opts.Tags = append(opts.Tags, tagStr)
			}
		}
	}
	
	// Handle domain filtering
	if domainsInterface, ok := args["domains"].([]interface{}); ok {
		for _, domain := range domainsInterface {
			if domainStr, ok := domain.(string); ok {
				opts.Domains = append(opts.Domains, domainStr)
			}
		}
	}

	// Perform find search
	results, err := s.cliTools.Find(ctx, opts)
	if err != nil {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32603,
				Message: "Find search failed",
				Data:    err.Error(),
			},
		}
	}

	// Format results in find style
	var output strings.Builder
	output.WriteString(fmt.Sprintf("Found %d items:\n\n", len(results)))
	
	for _, result := range results {
		output.WriteString(fmt.Sprintf("%s\n", result.Path))
		output.WriteString(fmt.Sprintf("  Type: %s, Size: %d bytes\n", result.ResourceType, result.Size))
		output.WriteString(fmt.Sprintf("  Created: %s, Modified: %s\n", 
			result.Created.Format("2006-01-02 15:04"), 
			result.Modified.Format("2006-01-02 15:04")))
		if len(result.Tags) > 0 {
			output.WriteString(fmt.Sprintf("  Tags: %s\n", strings.Join(result.Tags, ", ")))
		}
		output.WriteString("\n")
	}

	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      getValidID(request.ID),
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": output.String(),
				},
			},
		},
	}
}

// Handler for ripgrep-style search
func (s *Server) handleRipgrep(ctx context.Context, request *MCPRequest, args map[string]interface{}) *MCPResponse {
	pattern, ok := args["pattern"].(string)
	if !ok {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32602,
				Message: "Pattern is required",
			},
		}
	}

	// Build ripgrep options
	opts := &search.RipgrepOptions{
		Pattern: pattern,
	}

	// Apply optional parameters
	if ignoreCase, ok := args["ignore_case"].(bool); ok {
		opts.IgnoreCase = ignoreCase
	}
	if smartCase, ok := args["smart_case"].(bool); ok {
		opts.SmartCase = smartCase
	}
	if wordRegexp, ok := args["word_regexp"].(bool); ok {
		opts.WordRegexp = wordRegexp
	}
	if fixed, ok := args["fixed"].(bool); ok {
		opts.Fixed = fixed
	}
	if count, ok := args["count"].(bool); ok {
		opts.Count = count
	}
	if filesWithMatches, ok := args["files_with_matches"].(bool); ok {
		opts.FilesWithMatches = filesWithMatches
	}
	if lineNumber, ok := args["line_number"].(bool); ok {
		opts.LineNumber = lineNumber
	}
	if context, ok := args["context"].(float64); ok {
		opts.Context = int(context)
	}
	if maxCount, ok := args["max_count"].(float64); ok {
		opts.MaxCount = int(maxCount)
	}

	// Handle type filters
	if typeInterface, ok := args["type"].([]interface{}); ok {
		for _, t := range typeInterface {
			if typeStr, ok := t.(string); ok {
				opts.Type = append(opts.Type, typeStr)
			}
		}
	}
	
	// Handle domain filtering
	if domainsInterface, ok := args["domains"].([]interface{}); ok {
		for _, domain := range domainsInterface {
			if domainStr, ok := domain.(string); ok {
				opts.Domains = append(opts.Domains, domainStr)
			}
		}
	}

	// Perform ripgrep search
	results, err := s.cliTools.Ripgrep(ctx, opts)
	if err != nil {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32603,
				Message: "Ripgrep search failed",
				Data:    err.Error(),
			},
		}
	}

	// Format results in ripgrep style
	var output strings.Builder
	
	for _, result := range results {
		if opts.FilesWithMatches {
			output.WriteString(fmt.Sprintf("%s\n", result.Title))
		} else if opts.Count {
			output.WriteString(fmt.Sprintf("%s:%d\n", result.Title, result.MatchCount))
		} else {
			prefix := result.Title
			if opts.LineNumber && result.LineNumber > 0 {
				prefix = fmt.Sprintf("%s:%d", result.Title, result.LineNumber)
			}
			output.WriteString(fmt.Sprintf("%s:%s\n", prefix, result.MatchedLine))
			
			// Add context if available
			for _, contextLine := range result.Context {
				output.WriteString(fmt.Sprintf("%s\n", contextLine))
			}
		}
	}

	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      getValidID(request.ID),
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": output.String(),
				},
			},
		},
	}
}

// Domain management handlers

func (s *Server) handleCreateDomain(ctx context.Context, request *MCPRequest, args map[string]interface{}) *MCPResponse {
	id, ok := args["id"].(string)
	if !ok {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32602,
				Message: "Domain ID is required",
			},
		}
	}

	name, ok := args["name"].(string)
	if !ok {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32602,
				Message: "Domain name is required",
			},
		}
	}

	description, _ := args["description"].(string)
	parent, _ := args["parent"].(string)

	domain, err := s.domainManager.CreateDomain(ctx, id, name, description, parent)
	if err != nil {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32603,
				Message: "Failed to create domain",
				Data:    err.Error(),
			},
		}
	}

	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      getValidID(request.ID),
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": fmt.Sprintf("Domain '%s' created successfully with ID: %s\nPath: %s", domain.Name, domain.ID, domain.Path),
				},
			},
		},
	}
}

func (s *Server) handleListDomains(ctx context.Context, request *MCPRequest, args map[string]interface{}) *MCPResponse {
	parent, _ := args["parent"].(string)
	includeStats, _ := args["include_stats"].(bool)

	domains := s.domainManager.ListDomains(parent)

	var output strings.Builder
	output.WriteString(fmt.Sprintf("📁 Found %d domains:\n\n", len(domains)))

	for _, domain := range domains {
		output.WriteString(fmt.Sprintf("• %s (%s)\n", domain.Name, domain.ID))
		if domain.Description != "" {
			output.WriteString(fmt.Sprintf("  Description: %s\n", domain.Description))
		}
		output.WriteString(fmt.Sprintf("  Path: %s\n", domain.Path))
		if domain.Parent != "" {
			output.WriteString(fmt.Sprintf("  Parent: %s\n", domain.Parent))
		}
		output.WriteString(fmt.Sprintf("  Active: %t\n", domain.Active))
		output.WriteString(fmt.Sprintf("  Created: %s\n", domain.CreatedAt.Format("2006-01-02 15:04")))

		if includeStats {
			stats, err := s.storage.GetDomainStats(ctx, domain.ID)
			if err == nil {
				output.WriteString(fmt.Sprintf("  Resources: %d, Prompts: %d\n", stats["resources"], stats["prompts"]))
			}
		}
		output.WriteString("\n")
	}

	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      getValidID(request.ID),
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": output.String(),
				},
			},
		},
	}
}

func (s *Server) handleSwitchDomain(ctx context.Context, request *MCPRequest, args map[string]interface{}) *MCPResponse {
	domain, ok := args["domain"].(string)
	if !ok {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32602,
				Message: "Domain ID is required",
			},
		}
	}

	err := s.domainManager.SetCurrentDomain(domain)
	if err != nil {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32603,
				Message: "Failed to switch domain",
				Data:    err.Error(),
			},
		}
	}

	// Get domain info for confirmation
	domainInfo, err := s.domainManager.GetDomain(domain)
	if err != nil {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32603,
				Message: "Domain switched but failed to get info",
				Data:    err.Error(),
			},
		}
	}

	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      getValidID(request.ID),
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": fmt.Sprintf("✅ Switched to domain: %s (%s)\nPath: %s", domainInfo.Name, domainInfo.ID, domainInfo.Path),
				},
			},
		},
	}
}

func (s *Server) handleDomainStats(ctx context.Context, request *MCPRequest, args map[string]interface{}) *MCPResponse {
	domain, ok := args["domain"].(string)
	if !ok {
		domain = s.domainManager.GetCurrentDomain()
	}

	// Get domain info
	domainInfo, err := s.domainManager.GetDomain(domain)
	if err != nil {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32603,
				Message: "Domain not found",
				Data:    err.Error(),
			},
		}
	}

	// Get domain scope
	scope, err := s.domainManager.GetDomainScope(domain)
	if err != nil {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32603,
				Message: "Failed to get domain scope",
				Data:    err.Error(),
			},
		}
	}

	// Get storage stats
	stats, err := s.storage.GetDomainStats(ctx, domain)
	if err != nil {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32603,
				Message: "Failed to get domain statistics",
				Data:    err.Error(),
			},
		}
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("📊 Domain Statistics: %s (%s)\n\n", domainInfo.Name, domainInfo.ID))
	output.WriteString(fmt.Sprintf("Path: %s\n", domainInfo.Path))
	if domainInfo.Parent != "" {
		output.WriteString(fmt.Sprintf("Parent: %s\n", domainInfo.Parent))
	}
	output.WriteString(fmt.Sprintf("Active: %t\n", domainInfo.Active))
	output.WriteString(fmt.Sprintf("Created: %s\n", domainInfo.CreatedAt.Format("2006-01-02 15:04")))
	output.WriteString(fmt.Sprintf("Updated: %s\n\n", domainInfo.UpdatedAt.Format("2006-01-02 15:04")))

	output.WriteString("📦 Content Statistics:\n")
	output.WriteString(fmt.Sprintf("• Resources: %d\n", stats["resources"]))
	output.WriteString(fmt.Sprintf("• Prompts: %d\n\n", stats["prompts"]))

	output.WriteString("🔍 Domain Scope:\n")
	if len(scope.Ancestry) > 0 {
		output.WriteString(fmt.Sprintf("• Ancestry: %s\n", strings.Join(scope.Ancestry, " → ")))
	}
	if len(scope.Children) > 0 {
		output.WriteString(fmt.Sprintf("• Children: %s\n", strings.Join(scope.Children, ", ")))
	}
	output.WriteString(fmt.Sprintf("• Searchable domains: %s\n", strings.Join(scope.Searchable, ", ")))

	config := s.domainManager.GetConfig()
	output.WriteString(fmt.Sprintf("\n⚙️ Configuration:\n"))
	output.WriteString(fmt.Sprintf("• Isolation mode: %s\n", config.IsolationMode))
	output.WriteString(fmt.Sprintf("• Cross-domain access: %t\n", config.AllowCrossDomain))

	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      getValidID(request.ID),
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": output.String(),
				},
			},
		},
	}
}

// Shorthand notation resource and prompt handlers

func (s *Server) handleGetResource(ctx context.Context, request *MCPRequest, args map[string]interface{}) *MCPResponse {
	id, ok := args["id"].(string)
	if !ok {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32602,
				Message: "Resource ID is required",
			},
		}
	}

	// Parse domain and resource ID using shorthand notation
	domain, resourceID := s.domainManager.ParseResourceID(id)
	
	// Get resource from specific domain
	resource, err := s.storage.GetResourceInDomain(ctx, resourceID, domain)
	if err != nil {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32603,
				Message: "Resource not found",
				Data:    fmt.Sprintf("Failed to get resource %s from domain %s: %v", resourceID, domain, err),
			},
		}
	}

	// Format output with domain information
	var output strings.Builder
	output.WriteString(fmt.Sprintf("📄 Resource: %s\n", resource.Title))
	output.WriteString(fmt.Sprintf("🆔 ID: %s\n", s.domainManager.BuildResourceID(resource.Domain, resource.ID)))
	if resource.Domain != "" {
		output.WriteString(fmt.Sprintf("🏢 Domain: %s\n", resource.Domain))
	}
	if resource.Type != "" {
		output.WriteString(fmt.Sprintf("📋 Type: %s\n", resource.Type))
	}
	if len(resource.Tags) > 0 {
		output.WriteString(fmt.Sprintf("🏷️  Tags: %s\n", strings.Join(resource.Tags, ", ")))
	}
	output.WriteString(fmt.Sprintf("📅 Created: %s\n", resource.CreatedAt.Format("2006-01-02 15:04")))
	output.WriteString(fmt.Sprintf("📅 Updated: %s\n\n", resource.UpdatedAt.Format("2006-01-02 15:04")))
	output.WriteString("📝 Content:\n")
	output.WriteString(resource.Content)

	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      getValidID(request.ID),
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": output.String(),
				},
			},
		},
	}
}

func (s *Server) handleGetPrompt(ctx context.Context, request *MCPRequest, args map[string]interface{}) *MCPResponse {
	id, ok := args["id"].(string)
	if !ok {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32602,
				Message: "Prompt ID is required",
			},
		}
	}

	// Parse domain and prompt ID using shorthand notation
	domain, promptID := s.domainManager.ParseResourceID(id)
	
	// Get prompt from specific domain
	prompt, err := s.storage.GetPromptInDomain(ctx, promptID, domain)
	if err != nil {
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      getValidID(request.ID),
			Error: &MCPError{
				Code:    -32603,
				Message: "Prompt not found",
				Data:    fmt.Sprintf("Failed to get prompt %s from domain %s: %v", promptID, domain, err),
			},
		}
	}

	// Format output with domain information
	var output strings.Builder
	output.WriteString(fmt.Sprintf("📝 Prompt: %s\n", prompt.Name))
	output.WriteString(fmt.Sprintf("🆔 ID: %s\n", s.domainManager.BuildResourceID(prompt.Domain, prompt.ID)))
	if prompt.Domain != "" {
		output.WriteString(fmt.Sprintf("🏢 Domain: %s\n", prompt.Domain))
	}
	if prompt.Description != "" {
		output.WriteString(fmt.Sprintf("📋 Description: %s\n", prompt.Description))
	}
	if len(prompt.Variables) > 0 {
		output.WriteString("🔧 Variables:\n")
		for varName, varDesc := range prompt.Variables {
			output.WriteString(fmt.Sprintf("  • %s: %s\n", varName, varDesc))
		}
	}
	if len(prompt.Tags) > 0 {
		output.WriteString(fmt.Sprintf("🏷️  Tags: %s\n", strings.Join(prompt.Tags, ", ")))
	}
	output.WriteString(fmt.Sprintf("📅 Created: %s\n", prompt.CreatedAt.Format("2006-01-02 15:04")))
	output.WriteString(fmt.Sprintf("📅 Updated: %s\n\n", prompt.UpdatedAt.Format("2006-01-02 15:04")))
	output.WriteString("📄 Template:\n")
	output.WriteString(prompt.Template)

	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      getValidID(request.ID),
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": output.String(),
				},
			},
		},
	}
}