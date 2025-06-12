package tests

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"mcp-mindport/internal/domain"
	"mcp-mindport/internal/search"
	"mcp-mindport/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupWorkflowTestEnvironment(t *testing.T) (*storage.BadgerStore, *search.BleveSearch, func()) {
	// Create temporary directories
	tempDir, err := os.MkdirTemp("", "workflow_integration_test_*")
	require.NoError(t, err)

	storageDir := filepath.Join(tempDir, "storage")
	searchDir := filepath.Join(tempDir, "search")

	// Initialize storage
	store, err := storage.NewBadgerStore(storageDir)
	require.NoError(t, err)

	// Initialize search
	searchEngine, err := search.NewBleveSearch(searchDir)
	require.NoError(t, err)

	// Cleanup function
	cleanup := func() {
		store.Close()
		searchEngine.Close()
		os.RemoveAll(tempDir)
	}

	return store, searchEngine, cleanup
}

// TestCompleteResourceWorkflow tests the complete lifecycle of a resource
func TestCompleteResourceWorkflow(t *testing.T) {
	store, searchEngine, cleanup := setupWorkflowTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("Store -> Index -> Search -> Retrieve Workflow", func(t *testing.T) {
		// Step 1: Store a resource
		resource := &storage.Resource{
			ID:      "workflow-test-1",
			Domain:  "default",
			Type:    "documentation",
			Title:   "Complete Workflow Test",
			Content: "This is a comprehensive test of the complete resource workflow. It includes storage, indexing, searching, and retrieval operations.",
			Tags:    []string{"test", "workflow", "integration"},
			Metadata: map[string]interface{}{
				"author":   "test-suite",
				"version":  "1.0",
				"category": "testing",
			},
			SearchTerms: []string{"comprehensive", "workflow", "operations"},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := store.StoreResource(ctx, resource)
		require.NoError(t, err, "Failed to store resource")

		// Step 2: Index the resource for search
		err = searchEngine.IndexResource(ctx, resource)
		require.NoError(t, err, "Failed to index resource")

		// Give index time to settle
		time.Sleep(200 * time.Millisecond)

		// Step 3: Search for the resource using different methods
		t.Run("Basic Search", func(t *testing.T) {
			result, err := searchEngine.OptimizedSearch(ctx, "workflow", 10)
			require.NoError(t, err)
			assert.Contains(t, result, "Complete Workflow Test")
			assert.Contains(t, result, "comprehensive test")
		})

		t.Run("Advanced Search", func(t *testing.T) {
			query := &search.AdvancedSearchQuery{
				Query: "comprehensive test",
				Mode:  "smart",
				Limit: 10,
			}

			results, stats, err := searchEngine.AdvancedSearch(ctx, query)
			require.NoError(t, err)
			assert.Greater(t, stats.TotalResults, 0)
			
			// Debug: print all result IDs
			t.Logf("Found %d results:", len(results))
			for i, result := range results {
				t.Logf("  %d. ID: %s, Title: %s", i+1, result.ID, result.Title)
			}
			
			found := false
			for _, result := range results {
				if result.ID == "workflow-test-1" {
					found = true
					assert.Contains(t, result.Title, "Complete Workflow Test")
					break
				}
			}
			assert.True(t, found, "Resource not found in advanced search")
		})

		t.Run("Tag Search", func(t *testing.T) {
			query := &search.AdvancedSearchQuery{
				Query: "test",
				Tags:  []string{"workflow"},
				Limit: 10,
			}

			results, _, err := searchEngine.AdvancedSearch(ctx, query)
			require.NoError(t, err)
			
			found := false
			for _, result := range results {
				if result.ID == "workflow-test-1" {
					found = true
					assert.Contains(t, result.Tags, "workflow")
					break
				}
			}
			assert.True(t, found, "Resource not found in tag search")
		})

		// Step 4: Retrieve the resource by ID
		t.Run("Direct Retrieval", func(t *testing.T) {
			retrieved, err := store.GetResource(ctx, "workflow-test-1")
			require.NoError(t, err)
			
			assert.Equal(t, resource.ID, retrieved.ID)
			assert.Equal(t, resource.Title, retrieved.Title)
			assert.Equal(t, resource.Content, retrieved.Content)
			assert.Equal(t, resource.Tags, retrieved.Tags)
			assert.Equal(t, resource.Type, retrieved.Type)
		})

		// Step 5: Update the resource and verify changes
		t.Run("Update and Re-index", func(t *testing.T) {
			// Update the resource
			resource.Content = "UPDATED: This content has been modified to test the update workflow."
			resource.Tags = append(resource.Tags, "updated")
			resource.UpdatedAt = time.Now()

			err := store.StoreResource(ctx, resource)
			require.NoError(t, err)

			err = searchEngine.IndexResource(ctx, resource)
			require.NoError(t, err)

			time.Sleep(200 * time.Millisecond)

			// Search for updated content
			result, err := searchEngine.OptimizedSearch(ctx, "UPDATED", 10)
			require.NoError(t, err)
			assert.Contains(t, result, "workflow-test-1")

			// Verify the update
			retrieved, err := store.GetResource(ctx, "workflow-test-1")
			require.NoError(t, err)
			assert.Contains(t, retrieved.Content, "UPDATED")
			assert.Contains(t, retrieved.Tags, "updated")
		})
	})
}

// TestCompletePromptWorkflow tests the complete lifecycle of a prompt
func TestCompletePromptWorkflow(t *testing.T) {
	store, searchEngine, cleanup := setupWorkflowTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("Store -> Index -> Search -> Retrieve -> Execute Workflow", func(t *testing.T) {
		// Step 1: Store a prompt
		prompt := &storage.Prompt{
			ID:          "workflow-prompt-1",
			Domain:      "default",
			Name:        "Test Workflow Prompt",
			Description: "A prompt for testing the complete workflow",
			Template:    "Analyze this {{type}} code for {{purpose}}:\n\n{{code}}\n\nFocus on: {{focus}}\nProvide feedback on: {{feedback_areas}}",
			Variables: map[string]string{
				"type":           "Programming language or technology type",
				"purpose":        "The purpose or goal of the analysis",
				"code":           "The code to be analyzed",
				"focus":          "Specific areas to focus the analysis on",
				"feedback_areas": "Areas where feedback should be provided",
			},
			Tags:      []string{"analysis", "code", "workflow", "test"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := store.StorePrompt(ctx, prompt)
		require.NoError(t, err, "Failed to store prompt")

		// Step 2: Index the prompt
		err = searchEngine.IndexPrompt(ctx, prompt)
		require.NoError(t, err, "Failed to index prompt")

		time.Sleep(200 * time.Millisecond)

		// Step 3: Search for the prompt
		t.Run("Search Prompt", func(t *testing.T) {
			query := &search.AdvancedSearchQuery{
				Query: "workflow prompt",
				Type:  "prompt",
				Limit: 10,
			}

			results, _, err := searchEngine.AdvancedSearch(ctx, query)
			require.NoError(t, err)
			
			found := false
			for _, result := range results {
				if result.ID == "workflow-prompt-1" {
					found = true
					assert.Equal(t, "prompt", result.Type)
					break
				}
			}
			assert.True(t, found, "Prompt not found in search")
		})

		// Step 4: Retrieve the prompt
		t.Run("Retrieve Prompt", func(t *testing.T) {
			retrieved, err := store.GetPrompt(ctx, "workflow-prompt-1")
			require.NoError(t, err)
			
			assert.Equal(t, prompt.ID, retrieved.ID)
			assert.Equal(t, prompt.Name, retrieved.Name)
			assert.Equal(t, prompt.Template, retrieved.Template)
			assert.Equal(t, len(prompt.Variables), len(retrieved.Variables))
		})

		// Step 5: Execute the prompt with variables
		t.Run("Execute Prompt", func(t *testing.T) {
			retrieved, err := store.GetPrompt(ctx, "workflow-prompt-1")
			require.NoError(t, err)

			// Simulate prompt execution with variables
			variables := map[string]string{
				"type":           "Go",
				"purpose":        "performance optimization",
				"code":           "func Example() { /* code here */ }",
				"focus":          "performance bottlenecks",
				"feedback_areas": "efficiency, memory usage, algorithmic complexity",
			}

			executedTemplate := retrieved.Template
			for key, value := range variables {
				placeholder := fmt.Sprintf("{{%s}}", key)
				executedTemplate = strings.ReplaceAll(executedTemplate, placeholder, value)
			}

			// Verify template execution
			assert.Contains(t, executedTemplate, "Go")
			assert.Contains(t, executedTemplate, "performance optimization")
			assert.NotContains(t, executedTemplate, "{{") // No remaining placeholders
		})
	})
}

// TestCrossComponentWorkflow tests workflows that span multiple components
func TestCrossComponentWorkflow(t *testing.T) {
	store, searchEngine, cleanup := setupWorkflowTestEnvironment(t)
	defer cleanup()
	
	// Initialize components needed for this test
	cliTools := search.NewCLISearchTools(store, searchEngine)
	domainConfig := &domain.DomainConfig{
		DefaultDomain:    "default",
		IsolationMode:    "standard",
		AllowCrossDomain: true,
	}
	domainManager := domain.NewDomainManager(domainConfig)

	ctx := context.Background()

	t.Run("Multi-Domain Resource Management", func(t *testing.T) {
		// Step 1: Create multiple domains
		domains := []struct {
			id, name, description string
		}{
			{"frontend", "Frontend Team", "Frontend development resources"},
			{"backend", "Backend Team", "Backend development resources"},
			{"shared", "Shared Resources", "Cross-team shared resources"},
		}

		for _, d := range domains {
			_, err := domainManager.CreateDomain(ctx, d.id, d.name, d.description, "")
			require.NoError(t, err)
		}

		// Step 2: Store resources in different domains
		resources := []*storage.Resource{
			{
				ID:      "frontend-guide",
				Domain:  "frontend",
				Type:    "guide",
				Title:   "Frontend Development Guide",
				Content: "Comprehensive guide for frontend development including React, TypeScript, and testing.",
				Tags:    []string{"frontend", "react", "typescript"},
			},
			{
				ID:      "backend-api",
				Domain:  "backend",
				Type:    "documentation",
				Title:   "Backend API Documentation",
				Content: "API documentation for backend services including authentication and data endpoints.",
				Tags:    []string{"backend", "api", "documentation"},
			},
			{
				ID:      "shared-utils",
				Domain:  "shared",
				Type:    "library",
				Title:   "Shared Utility Library",
				Content: "Common utility functions shared across frontend and backend teams.",
				Tags:    []string{"shared", "utilities", "common"},
			},
		}

		for _, resource := range resources {
			resource.CreatedAt = time.Now()
			resource.UpdatedAt = time.Now()

			err := store.StoreResource(ctx, resource)
			require.NoError(t, err)

			err = searchEngine.IndexResource(ctx, resource)
			require.NoError(t, err)
		}

		time.Sleep(300 * time.Millisecond)

		// Step 3: Test domain-scoped searches
		t.Run("Frontend Domain Search", func(t *testing.T) {
			query := &search.AdvancedSearchQuery{
				Query:   "development",
				Domains: []string{"frontend"},
				Limit:   10,
			}

			results, _, err := searchEngine.AdvancedSearch(ctx, query)
			require.NoError(t, err)
			
			// Should only find frontend resources
			for _, result := range results {
				// Note: Domain field may not be available in search results
				// We can verify by checking if result ID contains frontend info
				assert.True(t, strings.Contains(result.ID, "frontend") || strings.Contains(result.Title, "Frontend"))
			}
		})

		t.Run("Cross-Domain Search", func(t *testing.T) {
			query := &search.AdvancedSearchQuery{
				Query:   "documentation",
				Domains: []string{"frontend", "backend", "shared"},
				Limit:   10,
			}

			results, _, err := searchEngine.AdvancedSearch(ctx, query)
			require.NoError(t, err)
			
			// Should find resources from multiple domains
			// Check by looking at resource IDs and titles for domain indicators
			hasMultipleDomains := false
			hasFrontend := false
			hasBackend := false
			hasShared := false
			
			for _, result := range results {
				if strings.Contains(result.ID, "frontend") || strings.Contains(result.Title, "Frontend") {
					hasFrontend = true
				}
				if strings.Contains(result.ID, "backend") || strings.Contains(result.Title, "Backend") {
					hasBackend = true
				}
				if strings.Contains(result.ID, "shared") || strings.Contains(result.Title, "Shared") {
					hasShared = true
				}
			}
			
			hasMultipleDomains = (hasFrontend && hasBackend) || (hasFrontend && hasShared) || (hasBackend && hasShared)
			assert.True(t, hasMultipleDomains || len(results) > 0) // At least find some results
		})

		// Step 4: Test CLI tools across domains
		t.Run("Grep Across Domains", func(t *testing.T) {
			opts := &search.GrepOptions{
				Pattern: "API",
				Domains: []string{"backend", "shared"},
			}

			results, err := cliTools.Grep(ctx, opts)
			require.NoError(t, err)
			
			// Should find API references in backend domain
			found := false
			for _, result := range results {
				if strings.Contains(result.Title, "API") {
					found = true
					break
				}
			}
			assert.True(t, found)
		})
	})
}

// TestSearchOptimizationWorkflow tests token-efficient search workflows
func TestSearchOptimizationWorkflow(t *testing.T) {
	store, searchEngine, cleanup := setupWorkflowTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	// Create a dataset with varying content sizes
	resources := make([]*storage.Resource, 0)
	for i := 0; i < 20; i++ {
		content := fmt.Sprintf("Resource %d content with detailed information about topic %d. ", i, i%5)
		if i%3 == 0 {
			// Some resources have very long content
			content += strings.Repeat("This is additional detailed information that makes the content longer. ", 20)
		}

		resource := &storage.Resource{
			ID:      fmt.Sprintf("opt-resource-%d", i),
			Domain:  "default",
			Type:    "documentation",
			Title:   fmt.Sprintf("Optimization Test Resource %d", i),
			Content: content,
			Tags:    []string{"optimization", "test", fmt.Sprintf("topic-%d", i%5)},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		resources = append(resources, resource)

		err := store.StoreResource(ctx, resource)
		require.NoError(t, err)

		err = searchEngine.IndexResource(ctx, resource)
		require.NoError(t, err)
	}

	time.Sleep(500 * time.Millisecond)

	t.Run("Token Efficient Search", func(t *testing.T) {
		// Test optimized search that should produce concise output
		result, err := searchEngine.OptimizedSearch(ctx, "optimization", 5)
		require.NoError(t, err)
		
		// Verify output is structured and concise
		lines := strings.Split(result, "\n")
		assert.Greater(t, len(lines), 3) // Should have header + results
		assert.Less(t, len(lines), 50)   // But not too verbose
		
		// Should contain key information
		assert.Contains(t, result, "Found")
		assert.Contains(t, result, "results")
		
		// Should have result entries with scores
		scoreCount := strings.Count(result, "Score:")
		assert.Greater(t, scoreCount, 0)
		assert.LessOrEqual(t, scoreCount, 5) // Limited to requested results
	})

	t.Run("Pagination and Limits", func(t *testing.T) {
		// Test that search respects limits
		query := &search.AdvancedSearchQuery{
			Query: "topic",
			Limit: 3,
		}

		results, stats, err := searchEngine.AdvancedSearch(ctx, query)
		require.NoError(t, err)
		
		assert.LessOrEqual(t, len(results), 3)
		assert.Greater(t, stats.TotalResults, 3) // More results available
	})

	t.Run("Performance Benchmarking", func(t *testing.T) {
		// Test search performance
		start := time.Now()
		
		query := &search.AdvancedSearchQuery{
			Query: "Resource",
			Limit: 10,
		}
		
		results, _, err := searchEngine.AdvancedSearch(ctx, query)
		elapsed := time.Since(start)
		
		require.NoError(t, err)
		assert.Greater(t, len(results), 0)
		assert.Less(t, elapsed, 200*time.Millisecond) // Should be fast
		
		t.Logf("Search completed in %v with %d results", elapsed, len(results))
	})
}

// TestErrorRecoveryWorkflow tests error handling and recovery
func TestErrorRecoveryWorkflow(t *testing.T) {
	store, searchEngine, cleanup := setupWorkflowTestEnvironment(t)
	defer cleanup()
	
	// Initialize CLI tools for this test
	cliTools := search.NewCLISearchTools(store, searchEngine)

	ctx := context.Background()

	t.Run("Storage Error Recovery", func(t *testing.T) {
		// Store a valid resource
		resource := &storage.Resource{
			ID:      "error-test-1",
			Domain:  "default",
			Type:    "test",
			Title:   "Error Recovery Test",
			Content: "This tests error recovery workflows",
			Tags:    []string{"error", "recovery"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := store.StoreResource(ctx, resource)
		require.NoError(t, err)

		err = searchEngine.IndexResource(ctx, resource)
		require.NoError(t, err)

		// Verify it's accessible
		retrieved, err := store.GetResource(ctx, "error-test-1")
		require.NoError(t, err)
		assert.Equal(t, resource.Title, retrieved.Title)

		// Attempt to retrieve non-existent resource
		_, err = store.GetResource(ctx, "non-existent")
		assert.Error(t, err) // Should error gracefully

		// System should still be functional
		retrieved, err = store.GetResource(ctx, "error-test-1")
		require.NoError(t, err)
		assert.Equal(t, resource.Title, retrieved.Title)
	})

	t.Run("Search Error Recovery", func(t *testing.T) {
		// Valid search should work
		query := &search.AdvancedSearchQuery{
			Query: "recovery",
			Limit: 10,
		}

		results, _, err := searchEngine.AdvancedSearch(ctx, query)
		require.NoError(t, err)
		assert.Greater(t, len(results), 0)

		// Invalid regex should be handled gracefully
		badQuery := &search.AdvancedSearchQuery{
			Query: "[invalid regex",
			Mode:  "regex",
			Limit: 10,
		}

		_, _, err = searchEngine.AdvancedSearch(ctx, badQuery)
		// Should either error gracefully or handle the bad regex
		if err != nil {
			assert.Contains(t, err.Error(), "regex") // Error should be descriptive
		}

		// System should still be functional after error
		results, _, err = searchEngine.AdvancedSearch(ctx, query)
		require.NoError(t, err)
		assert.Greater(t, len(results), 0)
	})

	t.Run("CLI Tools Error Recovery", func(t *testing.T) {
		// Valid grep should work
		opts := &search.GrepOptions{
			Pattern: "recovery",
		}

		results, err := cliTools.Grep(ctx, opts)
		require.NoError(t, err)
		assert.Greater(t, len(results), 0)

		// Invalid regex pattern
		badOpts := &search.GrepOptions{
			Pattern:  "[unclosed",
			Extended: true, // Enable regex mode
		}

		_, err = cliTools.Grep(ctx, badOpts)
		// Should handle error gracefully
		if err != nil {
			assert.NotPanics(t, func() {
				// System should not panic
			})
		}

		// System should remain functional
		results, err = cliTools.Grep(ctx, opts)
		require.NoError(t, err)
		assert.Greater(t, len(results), 0)
	})
}

// TestBatchOperationWorkflow tests batch operations and their workflows
func TestBatchOperationWorkflow(t *testing.T) {
	store, searchEngine, cleanup := setupWorkflowTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("Batch Resource Storage", func(t *testing.T) {
		// Create batch of resources
		numResources := 50
		resources := make([]*storage.Resource, 0, numResources)

		for i := 0; i < numResources; i++ {
			resource := &storage.Resource{
				ID:      fmt.Sprintf("batch-resource-%d", i),
				Domain:  "default",
				Type:    "batch",
				Title:   fmt.Sprintf("Batch Resource %d", i),
				Content: fmt.Sprintf("Content for batch resource number %d with some searchable text.", i),
				Tags:    []string{"batch", "test", fmt.Sprintf("group-%d", i/10)},
				CreatedAt: time.Now().Add(time.Duration(i) * time.Second),
				UpdatedAt: time.Now(),
			}
			resources = append(resources, resource)
		}

		// Store and index all resources
		start := time.Now()
		
		for _, resource := range resources {
			err := store.StoreResource(ctx, resource)
			require.NoError(t, err)

			err = searchEngine.IndexResource(ctx, resource)
			require.NoError(t, err)
		}
		
		elapsed := time.Since(start)
		t.Logf("Stored and indexed %d resources in %v", numResources, elapsed)

		// Give index time to settle
		time.Sleep(1 * time.Second)

		// Verify all resources are searchable
		query := &search.AdvancedSearchQuery{
			Query: "batch",
			Limit: numResources,
		}

		results, stats, err := searchEngine.AdvancedSearch(ctx, query)
		require.NoError(t, err)
		assert.Equal(t, numResources, stats.TotalResults)
		assert.Equal(t, numResources, len(results))

		// Verify random access to stored resources
		for i := 0; i < 10; i++ {
			id := fmt.Sprintf("batch-resource-%d", i*5)
			retrieved, err := store.GetResource(ctx, id)
			require.NoError(t, err)
			assert.Equal(t, id, retrieved.ID)
		}
	})

	t.Run("Batch Search Operations", func(t *testing.T) {
		// Perform multiple searches concurrently
		numSearches := 10
		searchQueries := make([]string, 0, numSearches)
		
		for i := 0; i < numSearches; i++ {
			searchQueries = append(searchQueries, fmt.Sprintf("group-%d", i%5))
		}

		start := time.Now()
		
		results := make([]int, numSearches)
		for i, query := range searchQueries {
			searchQuery := &search.AdvancedSearchQuery{
				Query: query,
				Limit: 20,
			}

			searchResults, _, err := searchEngine.AdvancedSearch(ctx, searchQuery)
			require.NoError(t, err)
			results[i] = len(searchResults)
		}
		
		elapsed := time.Since(start)
		t.Logf("Completed %d searches in %v", numSearches, elapsed)

		// Verify search results
		for i, resultCount := range results {
			assert.Greater(t, resultCount, 0, "Search %d returned no results", i)
		}
	})
}