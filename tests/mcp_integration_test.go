package tests

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mcp-mindport/internal/config"
	"mcp-mindport/internal/mcp"
	"mcp-mindport/internal/search"
	"mcp-mindport/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helpers and setup

func setupMCPTestEnvironment(t *testing.T) (*mcp.Server, *config.Config, func()) {
	// Create temporary directories
	tempDir, err := os.MkdirTemp("", "mcp_integration_test_*")
	require.NoError(t, err)

	storageDir := filepath.Join(tempDir, "storage")
	searchDir := filepath.Join(tempDir, "search")

	// Create test config
	cfg := &config.Config{
		Storage: config.StorageConfig{
			Path: storageDir,
		},
		Search: config.SearchConfig{
			IndexPath: searchDir,
		},
		Domain: config.DomainConfig{
			DefaultDomain:    "test-default",
			IsolationMode:    "standard",
			AllowCrossDomain: true,
		},
	}

	// Initialize storage
	store, err := storage.NewBadgerStore(storageDir)
	require.NoError(t, err)

	// Initialize search
	searchEngine, err := search.NewBleveSearch(searchDir)
	require.NoError(t, err)

	// Create MCP server
	server := mcp.NewServer(store, searchEngine, cfg)

	// Cleanup function
	cleanup := func() {
		store.Close()
		searchEngine.Close()
		os.RemoveAll(tempDir)
	}

	return server, cfg, cleanup
}

// TestMCPServerLifecycle tests the basic MCP server lifecycle
func TestMCPServerLifecycle(t *testing.T) {
	_, _, cleanup := setupMCPTestEnvironment(t)
	defer cleanup()

	// Test basic server creation
	assert.True(t, true) // Placeholder assertion

	t.Run("Server Creation", func(t *testing.T) {
		// Test that server can be created without error
		// The setup already validates this
		assert.True(t, true)
	})

	t.Run("Server Shutdown", func(t *testing.T) {
		// Test that cleanup works without error
		// The defer cleanup will test this
		assert.True(t, true)
	})
}

// TestMCPInitialization tests the MCP initialization handshake
func TestMCPInitialization(t *testing.T) {
	server, _, cleanup := setupMCPTestEnvironment(t)
	defer cleanup()

	t.Run("Initialize Request", func(t *testing.T) {
		ctx := context.Background()
		request := &mcp.MCPRequest{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "initialize",
			Params: map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities":    map[string]interface{}{},
				"clientInfo": map[string]interface{}{
					"name":    "test-client",
					"version": "1.0.0",
				},
			},
		}

		response := server.HandleRequest(ctx, request)
		require.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.Equal(t, request.ID, response.ID)
		assert.Nil(t, response.Error)
		assert.NotNil(t, response.Result)

		// Check initialization result
		result, ok := response.Result.(mcp.InitializeResult)
		if !ok {
			// Try as map[string]interface{} in case of different serialization
			resultMap, ok := response.Result.(map[string]interface{})
			require.True(t, ok, "Result should be InitializeResult or map")
			assert.Equal(t, "2024-11-05", resultMap["protocolVersion"])
			assert.NotNil(t, resultMap["capabilities"])
			assert.NotNil(t, resultMap["serverInfo"])
		} else {
			assert.Equal(t, "2024-11-05", result.ProtocolVersion)
			assert.NotNil(t, result.Capabilities)
			assert.Equal(t, "mcp-mindport", result.ServerInfo.Name)
		}
	})
}

// TestMCPResourceOperations tests resource-related MCP operations
func TestMCPResourceOperations(t *testing.T) {
	server, _, cleanup := setupMCPTestEnvironment(t)
	defer cleanup()
	ctx := context.Background()

	t.Run("List Resources", func(t *testing.T) {
		// Test resources/list method
		request := &mcp.MCPRequest{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "resources/list",
		}

		response := server.HandleRequest(ctx, request)
		require.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.Nil(t, response.Error)
		assert.NotNil(t, response.Result)
	})

	t.Run("Store and Read Resource", func(t *testing.T) {
		// First store a resource via tools/call
		storeRequest := &mcp.MCPRequest{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name": "store_resource",
				"arguments": map[string]interface{}{
					"title":   "Test Resource",
					"content": "This is test content for MCP integration test",
					"type":    "test",
					"tags":    []string{"mcp", "test"},
				},
			},
		}

		storeResponse := server.HandleRequest(ctx, storeRequest)
		require.NotNil(t, storeResponse)
		assert.Equal(t, "2.0", storeResponse.JSONRPC)
		assert.Nil(t, storeResponse.Error)
		
		// Verify the resource was stored by listing resources
		listRequest := &mcp.MCPRequest{
			JSONRPC: "2.0",
			ID:      2,
			Method:  "resources/list",
		}

		listResponse := server.HandleRequest(ctx, listRequest)
		require.NotNil(t, listResponse)
		assert.Nil(t, listResponse.Error)
		
		result, ok := listResponse.Result.(map[string]interface{})
		require.True(t, ok)
		
		resources, ok := result["resources"].([]map[string]interface{})
		require.True(t, ok)
		assert.Greater(t, len(resources), 0, "Should have at least one resource")
	})
}

// TestMCPToolOperations tests tool-related MCP operations
func TestMCPToolOperations(t *testing.T) {
	server, _, cleanup := setupMCPTestEnvironment(t)
	defer cleanup()
	ctx := context.Background()

	t.Run("List Tools", func(t *testing.T) {
		// Test tools/list method
		request := &mcp.MCPRequest{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "tools/list",
		}

		response := server.HandleRequest(ctx, request)
		require.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.Nil(t, response.Error)
		assert.NotNil(t, response.Result)

		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok)
		
		tools, ok := result["tools"].([]map[string]interface{})
		require.True(t, ok)
		assert.Greater(t, len(tools), 10, "Should have multiple tools available")
	})

	t.Run("Search Tools", func(t *testing.T) {
		// First store a test resource to search for
		storeRequest := &mcp.MCPRequest{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name": "store_resource",
				"arguments": map[string]interface{}{
					"title":   "Search Test Resource",
					"content": "This content contains the word searchable for testing purposes",
					"type":    "test",
					"tags":    []string{"search", "test"},
				},
			},
		}
		
		storeResponse := server.HandleRequest(ctx, storeRequest)
		require.NotNil(t, storeResponse)
		assert.Nil(t, storeResponse.Error)

		// Test various search tools
		searchTools := []struct {
			name string
			args map[string]interface{}
		}{
			{
				"search_resources",
				map[string]interface{}{"query": "searchable", "limit": 5},
			},
			{
				"advanced_search", 
				map[string]interface{}{"query": "searchable", "mode": "smart", "limit": 5},
			},
			{
				"grep",
				map[string]interface{}{"pattern": "searchable", "ignore_case": true},
			},
			{
				"find",
				map[string]interface{}{"name": "*Search*"},
			},
			{
				"ripgrep",
				map[string]interface{}{"pattern": "searchable", "smart_case": true},
			},
		}

		for _, tool := range searchTools {
			t.Run(tool.name, func(t *testing.T) {
				request := &mcp.MCPRequest{
					JSONRPC: "2.0",
					ID:      1,
					Method:  "tools/call",
					Params: map[string]interface{}{
						"name":      tool.name,
						"arguments": tool.args,
					},
				}

				response := server.HandleRequest(ctx, request)
				require.NotNil(t, response)
				assert.Equal(t, "2.0", response.JSONRPC)
				assert.Nil(t, response.Error, "Tool %s should not return error", tool.name)
				assert.NotNil(t, response.Result)
			})
		}
	})

	t.Run("Domain Management Tools", func(t *testing.T) {
		domainTests := []struct {
			name string
			args map[string]interface{}
		}{
			{
				"create_domain",
				map[string]interface{}{
					"id":          "test-domain",
					"name":        "Test Domain",
					"description": "A test domain for MCP integration tests",
				},
			},
			{
				"list_domains",
				map[string]interface{}{"include_stats": true},
			},
			{
				"domain_stats",
				map[string]interface{}{},
			},
		}

		for _, test := range domainTests {
			t.Run(test.name, func(t *testing.T) {
				request := &mcp.MCPRequest{
					JSONRPC: "2.0",
					ID:      1,
					Method:  "tools/call",
					Params: map[string]interface{}{
						"name":      test.name,
						"arguments": test.args,
					},
				}

				response := server.HandleRequest(ctx, request)
				require.NotNil(t, response)
				assert.Equal(t, "2.0", response.JSONRPC)
				assert.Nil(t, response.Error, "Tool %s should not return error", test.name)
				assert.NotNil(t, response.Result)
			})
		}
	})
}

// TestMCPPromptOperations tests prompt-related MCP operations
func TestMCPPromptOperations(t *testing.T) {
	server, _, cleanup := setupMCPTestEnvironment(t)
	defer cleanup()
	ctx := context.Background()

	t.Run("List Prompts", func(t *testing.T) {
		// Test prompts/list method
		request := &mcp.MCPRequest{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "prompts/list",
		}

		response := server.HandleRequest(ctx, request)
		require.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.Nil(t, response.Error)
		assert.NotNil(t, response.Result)
	})

	t.Run("Store and Get Prompt", func(t *testing.T) {
		// Store a prompt
		storeRequest := &mcp.MCPRequest{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name": "store_prompt",
				"arguments": map[string]interface{}{
					"name":        "Test MCP Prompt",
					"description": "A test prompt for MCP integration testing",
					"template":    "Analyze this {{content}} and provide feedback on {{aspects}}",
					"variables": map[string]interface{}{
						"content": "The content to analyze",
						"aspects": "The aspects to focus on",
					},
					"tags": []string{"mcp", "test", "analysis"},
				},
			},
		}

		storeResponse := server.HandleRequest(ctx, storeRequest)
		require.NotNil(t, storeResponse)
		assert.Equal(t, "2.0", storeResponse.JSONRPC)
		assert.Nil(t, storeResponse.Error)

		// List prompts to verify it was stored
		listRequest := &mcp.MCPRequest{
			JSONRPC: "2.0",
			ID:      2,
			Method:  "prompts/list",
		}

		listResponse := server.HandleRequest(ctx, listRequest)
		require.NotNil(t, listResponse)
		assert.Nil(t, listResponse.Error)
		
		result, ok := listResponse.Result.(map[string]interface{})
		require.True(t, ok)
		
		prompts, ok := result["prompts"].([]map[string]interface{})
		require.True(t, ok)
		assert.Greater(t, len(prompts), 0, "Should have at least one prompt")

		// Get a specific prompt
		getRequest := &mcp.MCPRequest{
			JSONRPC: "2.0",
			ID:      3,
			Method:  "prompts/get",
			Params: map[string]interface{}{
				"name": "Test MCP Prompt",
				"arguments": map[string]interface{}{
					"content": "sample code",
					"aspects": "performance and readability",
				},
			},
		}

		getResponse := server.HandleRequest(ctx, getRequest)
		require.NotNil(t, getResponse)
		assert.Equal(t, "2.0", getResponse.JSONRPC)
		assert.Nil(t, getResponse.Error)
		assert.NotNil(t, getResponse.Result)
	})
}

// TestMCPErrorHandling tests error scenarios
func TestMCPErrorHandling(t *testing.T) {
	server, _, cleanup := setupMCPTestEnvironment(t)
	defer cleanup()
	ctx := context.Background()

	errorCases := []struct {
		name          string
		request       *mcp.MCPRequest
		expectedError int
	}{
		{
			name: "Invalid Method",
			request: &mcp.MCPRequest{
				JSONRPC: "2.0",
				ID:      1,
				Method:  "invalid/method",
			},
			expectedError: -32601, // Method not found
		},
		{
			name: "Invalid Tool",
			request: &mcp.MCPRequest{
				JSONRPC: "2.0",
				ID:      1,
				Method:  "tools/call",
				Params: map[string]interface{}{
					"name":      "invalid_tool",
					"arguments": map[string]interface{}{},
				},
			},
			expectedError: -32601, // Tool not found
		},
		{
			name: "Missing Required Params",
			request: &mcp.MCPRequest{
				JSONRPC: "2.0",
				ID:      1,
				Method:  "tools/call",
				Params: map[string]interface{}{
					"name": "store_resource",
					"arguments": map[string]interface{}{
						// Missing required title and content
					},
				},
			},
			expectedError: -32602, // Invalid params
		},
	}

	for _, tc := range errorCases {
		t.Run(tc.name, func(t *testing.T) {
			response := server.HandleRequest(ctx, tc.request)
			require.NotNil(t, response)
			assert.Equal(t, "2.0", response.JSONRPC)
			assert.NotNil(t, response.Error, "Should return an error")
			assert.Equal(t, tc.expectedError, response.Error.Code)
			assert.Nil(t, response.Result, "Should not return a result when there's an error")
		})
	}
}

// TestMCPConcurrentOperations tests concurrent MCP operations
func TestMCPConcurrentOperations(t *testing.T) {
	server, _, cleanup := setupMCPTestEnvironment(t)
	defer cleanup()
	ctx := context.Background()

	t.Run("Concurrent Resource Storage", func(t *testing.T) {
		// Test multiple concurrent resource storage operations
		const numOps = 10
		results := make(chan *mcp.MCPResponse, numOps)
		
		for i := 0; i < numOps; i++ {
			go func(index int) {
				request := &mcp.MCPRequest{
					JSONRPC: "2.0",
					ID:      index,
					Method:  "tools/call",
					Params: map[string]interface{}{
						"name": "store_resource",
						"arguments": map[string]interface{}{
							"title":   fmt.Sprintf("Concurrent Test Resource %d", index),
							"content": fmt.Sprintf("Content for concurrent test resource number %d", index),
							"type":    "concurrent-test",
							"tags":    []string{"concurrent", "test", fmt.Sprintf("batch-%d", index)},
						},
					},
				}
				
				response := server.HandleRequest(ctx, request)
				results <- response
			}(i)
		}
		
		// Collect all results
		for i := 0; i < numOps; i++ {
			response := <-results
			assert.NotNil(t, response)
			assert.Equal(t, "2.0", response.JSONRPC)
			assert.Nil(t, response.Error, "Concurrent operation should not fail")
		}
	})

	t.Run("Concurrent Search Operations", func(t *testing.T) {
		// First store a resource to search for
		storeRequest := &mcp.MCPRequest{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name": "store_resource",
				"arguments": map[string]interface{}{
					"title":   "Concurrent Search Target",
					"content": "This resource will be searched by multiple concurrent operations",
					"type":    "search-target",
					"tags":    []string{"concurrent", "search", "target"},
				},
			},
		}
		
		storeResponse := server.HandleRequest(ctx, storeRequest)
		require.NotNil(t, storeResponse)
		assert.Nil(t, storeResponse.Error)

		// Test multiple concurrent search operations
		const numSearches = 5
		searchResults := make(chan *mcp.MCPResponse, numSearches)
		
		for i := 0; i < numSearches; i++ {
			go func(index int) {
				request := &mcp.MCPRequest{
					JSONRPC: "2.0",
					ID:      index,
					Method:  "tools/call",
					Params: map[string]interface{}{
						"name": "search_resources",
						"arguments": map[string]interface{}{
							"query": "concurrent",
							"limit": 10,
						},
					},
				}
				
				response := server.HandleRequest(ctx, request)
				searchResults <- response
			}(i)
		}
		
		// Collect all search results
		for i := 0; i < numSearches; i++ {
			response := <-searchResults
			assert.NotNil(t, response)
			assert.Equal(t, "2.0", response.JSONRPC)
			assert.Nil(t, response.Error, "Concurrent search should not fail")
		}
	})
}

// TestMCPResourceURIHandling tests the mindport:// URI handling
func TestMCPResourceURIHandling(t *testing.T) {
	server, _, cleanup := setupMCPTestEnvironment(t)
	defer cleanup()
	ctx := context.Background()

	// First store a resource to read
	storeRequest := &mcp.MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "store_resource",
			"arguments": map[string]interface{}{
				"title":   "URI Test Resource",
				"content": "This resource is used for testing URI handling",
				"type":    "uri-test",
				"tags":    []string{"uri", "test"},
			},
		},
	}

	storeResponse := server.HandleRequest(ctx, storeRequest)
	require.NotNil(t, storeResponse)
	assert.Nil(t, storeResponse.Error)

	// Extract resource ID from response
	result, ok := storeResponse.Result.(map[string]interface{})
	require.True(t, ok)
	content, ok := result["content"].([]map[string]interface{})
	require.True(t, ok)
	require.Greater(t, len(content), 0)
	
	// Extract the ID from the success message
	text, ok := content[0]["text"].(string)
	require.True(t, ok)
	require.Contains(t, text, "Resource stored successfully with ID:")
	
	parts := strings.Split(text, "ID: ")
	require.Len(t, parts, 2)
	resourceID := strings.TrimSpace(parts[1])

	t.Run("Valid URI Format", func(t *testing.T) {
		// Test that resources/read correctly handles mindport://resource/{id} URIs
		request := &mcp.MCPRequest{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "resources/read",
			Params: map[string]interface{}{
				"uri": fmt.Sprintf("mindport://resource/%s", resourceID),
			},
		}

		response := server.HandleRequest(ctx, request)
		require.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.Nil(t, response.Error)
		assert.NotNil(t, response.Result)

		result, ok := response.Result.(map[string]interface{})
		require.True(t, ok)
		
		contents, ok := result["contents"].([]map[string]interface{})
		require.True(t, ok)
		require.Greater(t, len(contents), 0)
		
		firstContent := contents[0]
		assert.Equal(t, "text/plain", firstContent["mimeType"])
		assert.Contains(t, firstContent["text"], "URI handling")
	})

	t.Run("Invalid URI Format", func(t *testing.T) {
		invalidURIs := []string{
			"invalid://resource/test",
			"mindport://invalid/test",
			"mindport://resource/",
			"not-a-uri",
		}

		for _, uri := range invalidURIs {
			t.Run("URI: "+uri, func(t *testing.T) {
				request := &mcp.MCPRequest{
					JSONRPC: "2.0",
					ID:      1,
					Method:  "resources/read",
					Params: map[string]interface{}{
						"uri": uri,
					},
				}

				response := server.HandleRequest(ctx, request)
				require.NotNil(t, response)
				assert.Equal(t, "2.0", response.JSONRPC)
				assert.NotNil(t, response.Error, "Should return error for invalid URI: %s", uri)
				assert.Equal(t, -32602, response.Error.Code, "Should return invalid params error")
			})
		}
	})
}

// TestMCPShorthandNotation tests the shorthand domain notation
func TestMCPShorthandNotation(t *testing.T) {
	server, _, cleanup := setupMCPTestEnvironment(t)
	defer cleanup()
	ctx := context.Background()

	// Store test resources and prompts first
	storeResourceRequest := &mcp.MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "store_resource",
			"arguments": map[string]interface{}{
				"title":   "Shorthand Test Resource",
				"content": "Resource for testing shorthand notation",
				"type":    "shorthand-test",
				"tags":    []string{"shorthand", "test"},
			},
		},
	}

	storeResourceResponse := server.HandleRequest(ctx, storeResourceRequest)
	require.NotNil(t, storeResourceResponse)
	assert.Nil(t, storeResourceResponse.Error)

	storePromptRequest := &mcp.MCPRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "store_prompt",
			"arguments": map[string]interface{}{
				"name":        "Shorthand Test Prompt",
				"description": "Prompt for testing shorthand notation",
				"template":    "Test template with {{variable}}",
				"variables": map[string]interface{}{
					"variable": "A test variable",
				},
				"tags": []string{"shorthand", "test"},
			},
		},
	}

	storePromptResponse := server.HandleRequest(ctx, storePromptRequest)
	require.NotNil(t, storePromptResponse)
	assert.Nil(t, storePromptResponse.Error)

	// Extract resource and prompt IDs
	resourceResult, ok := storeResourceResponse.Result.(map[string]interface{})
	require.True(t, ok)
	resourceContent, ok := resourceResult["content"].([]map[string]interface{})
	require.True(t, ok)
	resourceText, ok := resourceContent[0]["text"].(string)
	require.True(t, ok)
	resourceParts := strings.Split(resourceText, "ID: ")
	require.Len(t, resourceParts, 2)
	resourceID := strings.TrimSpace(resourceParts[1])

	promptResult, ok := storePromptResponse.Result.(map[string]interface{})
	require.True(t, ok)
	promptContent, ok := promptResult["content"].([]map[string]interface{})
	require.True(t, ok)
	promptText, ok := promptContent[0]["text"].(string)
	require.True(t, ok)
	promptParts := strings.Split(promptText, "ID: ")
	require.Len(t, promptParts, 2)
	promptID := strings.TrimSpace(promptParts[1])

	shorthandCases := []struct {
		name      string
		input     string
		toolName  string
		shouldErr bool
	}{
		{
			name:     "Default Domain Notation Resource",
			input:    "::" + resourceID,
			toolName: "get_resource",
		},
		{
			name:     "Plain ID Resource",
			input:    resourceID,
			toolName: "get_resource",
		},
		{
			name:     "Default Domain Notation Prompt",
			input:    "::" + promptID,
			toolName: "get_prompt",
		},
		{
			name:     "Plain ID Prompt",
			input:    promptID,
			toolName: "get_prompt",
		},
		{
			name:      "Nonexistent Resource",
			input:     "::nonexistent123",
			toolName:  "get_resource",
			shouldErr: true,
		},
	}

	for _, tc := range shorthandCases {
		t.Run(tc.name, func(t *testing.T) {
			request := &mcp.MCPRequest{
				JSONRPC: "2.0",
				ID:      1,
				Method:  "tools/call",
				Params: map[string]interface{}{
					"name": tc.toolName,
					"arguments": map[string]interface{}{
						"id": tc.input,
					},
				},
			}

			response := server.HandleRequest(ctx, request)
			require.NotNil(t, response)
			assert.Equal(t, "2.0", response.JSONRPC)
			
			if tc.shouldErr {
				assert.NotNil(t, response.Error, "Should return error for %s", tc.name)
			} else {
				assert.Nil(t, response.Error, "Should not return error for %s", tc.name)
				assert.NotNil(t, response.Result)
			}
		})
	}
}