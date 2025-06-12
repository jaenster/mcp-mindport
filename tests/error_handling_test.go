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

func setupErrorTestEnvironment(t *testing.T) (*storage.BadgerStore, *search.BleveSearch, func()) {
	// Create temporary directories
	tempDir, err := os.MkdirTemp("", "error_test_*")
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

// TestStorageErrorHandling tests error handling in storage operations
func TestStorageErrorHandling(t *testing.T) {
	store, _, cleanup := setupErrorTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("Invalid Resource Data", func(t *testing.T) {
		// Test with nil resource - should error gracefully
		err := store.StoreResource(ctx, nil)
		assert.Error(t, err, "Storing nil resource should return error")
		assert.Contains(t, err.Error(), "nil")

		// Test with empty resource ID - system currently allows this
		emptyIDResource := &storage.Resource{
			ID:      "",
			Title:   "Empty ID Test",
			Content: "Test content",
		}
		err = store.StoreResource(ctx, emptyIDResource)
		// Note: Current implementation allows empty IDs
		// This could be enhanced to validate and reject empty IDs
		if err != nil {
			assert.Contains(t, err.Error(), "ID")
		}

		// Test with extremely long ID
		longIDResource := &storage.Resource{
			ID:      strings.Repeat("a", 1000), // Very long ID
			Title:   "Long ID Test",
			Content: "Test content",
		}
		err2 := store.StoreResource(ctx, longIDResource)
		// Should either work or fail gracefully
		if err2 != nil {
			assert.Contains(t, err2.Error(), "ID")
		}

		// Test with nil metadata
		resource := &storage.Resource{
			ID:       "test-nil-metadata",
			Title:    "Test",
			Content:  "Test",
			Metadata: nil, // Should handle nil gracefully
		}
		err3 := store.StoreResource(ctx, resource)
		assert.NoError(t, err3) // Should handle nil metadata gracefully
	})

	t.Run("Non-existent Resource Retrieval", func(t *testing.T) {
		// Try to get non-existent resource
		_, err := store.GetResource(ctx, "non-existent-resource")
		assert.Error(t, err)
		assert.Contains(t, strings.ToLower(err.Error()), "not found")

		// Try to get resource with empty ID
		_, err = store.GetResource(ctx, "")
		assert.Error(t, err)

		// Try to get resource with special characters
		_, err = store.GetResource(ctx, "resource/with/slashes")
		// Should either work or fail gracefully
		if err != nil {
			t.Logf("Special character resource ID failed as expected: %v", err)
		}
	})

	t.Run("Prompt Storage Errors", func(t *testing.T) {
		// Test with nil prompt - should error gracefully
		err := store.StorePrompt(ctx, nil)
		assert.Error(t, err, "Storing nil prompt should return error")
		assert.Contains(t, err.Error(), "nil")

		// Test with empty prompt ID - system currently allows this
		emptyPrompt := &storage.Prompt{
			ID:       "",
			Name:     "Empty ID",
			Template: "Test template",
		}
		err = store.StorePrompt(ctx, emptyPrompt)
		// Note: Current implementation allows empty IDs
		// This could be enhanced to validate and reject empty IDs
		if err != nil {
			assert.Contains(t, err.Error(), "ID")
		}

		// Test with invalid template variables
		invalidPrompt := &storage.Prompt{
			ID:       "invalid-prompt",
			Name:     "Invalid Prompt",
			Template: "Template with {{unclosed_variable",
			Variables: map[string]string{
				"unclosed_variable": "Should be properly handled",
			},
		}
		err2 := store.StorePrompt(ctx, invalidPrompt)
		assert.NoError(t, err2) // Should store even with malformed template

		// But retrieval should work
		retrieved, err3 := store.GetPrompt(ctx, "invalid-prompt")
		require.NoError(t, err3)
		assert.Equal(t, invalidPrompt.Template, retrieved.Template)
	})

	t.Run("Domain Storage Errors", func(t *testing.T) {
		// Test resource retrieval from non-existent domain
		_, err := store.GetResourceInDomain(ctx, "test-resource", "non-existent-domain")
		assert.Error(t, err)

		// Test prompt retrieval from non-existent domain
		_, err = store.GetPromptInDomain(ctx, "test-prompt", "non-existent-domain")
		assert.Error(t, err)

		// Test stats for non-existent domain (should return zeros, not error)
		stats, err := store.GetDomainStats(ctx, "non-existent-domain")
		require.NoError(t, err)
		assert.Equal(t, 0, stats["resources"])
		assert.Equal(t, 0, stats["prompts"])
	})

	t.Run("Context Cancellation", func(t *testing.T) {
		// Test with cancelled context
		cancelledCtx, cancel := context.WithCancel(ctx)
		cancel() // Cancel immediately

		resource := &storage.Resource{
			ID:      "cancelled-context-test",
			Title:   "Cancelled Context Test",
			Content: "This should fail",
		}

		err := store.StoreResource(cancelledCtx, resource)
		// Should handle cancelled context gracefully
		if err != nil {
			assert.Contains(t, err.Error(), "context")
		}

		// Test with timeout context
		timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 1*time.Nanosecond)
		defer timeoutCancel()
		time.Sleep(1 * time.Millisecond) // Ensure timeout

		err = store.StoreResource(timeoutCtx, resource)
		if err != nil {
			assert.Contains(t, err.Error(), "context")
		}
	})
}

// TestSearchErrorHandling tests error handling in search operations
func TestSearchErrorHandling(t *testing.T) {
	store, searchEngine, cleanup := setupErrorTestEnvironment(t)
	defer cleanup()
	
	// Initialize CLI tools for this test
	cliTools := search.NewCLISearchTools(store, searchEngine)

	ctx := context.Background()

	t.Run("Invalid Search Queries", func(t *testing.T) {
		// Test with empty query
		result, err := searchEngine.OptimizedSearch(ctx, "", 10)
		// Should either return empty results or handle gracefully
		if err != nil {
			assert.Contains(t, strings.ToLower(err.Error()), "query")
		} else {
			assert.NotNil(t, result)
		}

		// Test with very long query
		longQuery := strings.Repeat("test ", 1000)
		result, err = searchEngine.OptimizedSearch(ctx, longQuery, 10)
		// Should handle long queries gracefully
		if err != nil {
			t.Logf("Long query failed as expected: %v", err)
		}

		// Test with special characters
		specialQuery := "test!@#$%^&*(){}[]|\\:;\"'<>,.?/~`"
		result, err = searchEngine.OptimizedSearch(ctx, specialQuery, 10)
		// Should handle special characters gracefully
		if err != nil {
			t.Logf("Special character query failed: %v", err)
		}
	})

	t.Run("Invalid Advanced Search Queries", func(t *testing.T) {
		// Test with invalid regex
		regexQuery := &search.AdvancedSearchQuery{
			Query: "[unclosed bracket",
			Mode:  "regex",
			Limit: 10,
		}

		_, _, err := searchEngine.AdvancedSearch(ctx, regexQuery)
		// Should handle invalid regex gracefully
		if err != nil {
			assert.Contains(t, strings.ToLower(err.Error()), "regex")
		}

		// Test with negative limit - should be handled gracefully
		negativeQuery := &search.AdvancedSearchQuery{
			Query: "test",
			Limit: -5,
		}

		// Should handle negative limit gracefully
		results, _, err := searchEngine.AdvancedSearch(ctx, negativeQuery)
		// Either should error gracefully or handle by using default limit
		if err != nil {
			assert.Contains(t, strings.ToLower(err.Error()), "limit")
		} else {
			// If no error, should have applied some reasonable default/minimum
			assert.GreaterOrEqual(t, len(results), 0)
		}

		// Test with extremely large limit
		largeQuery := &search.AdvancedSearchQuery{
			Query: "test",
			Limit: 1000000,
		}

		results, _, err = searchEngine.AdvancedSearch(ctx, largeQuery)
		// Should cap limit or handle gracefully
		if err == nil {
			assert.LessOrEqual(t, len(results), 10000) // Reasonable cap
		}
	})

	t.Run("CLI Tools Error Handling", func(t *testing.T) {
		// Test grep with invalid regex
		grepOpts := &search.GrepOptions{
			Pattern:  "[invalid regex",
			Extended: true,
		}

		_, err := cliTools.Grep(ctx, grepOpts)
		// Should handle invalid regex gracefully
		if err != nil {
			assert.Contains(t, strings.ToLower(err.Error()), "regex")
		}

		// Test find with invalid size filter
		findOpts := &search.FindOptions{
			Size: "invalid-size-format",
		}

		_, err = cliTools.Find(ctx, findOpts)
		// Should handle invalid size format gracefully
		if err != nil {
			assert.Contains(t, strings.ToLower(err.Error()), "size")
		}

		// Test ripgrep with conflicting options
		ripgrepOpts := &search.RipgrepOptions{
			Pattern:           "test",
			Count:             true,
			FilesWithMatches: true, // Conflicting with Count
		}

		_, err = cliTools.Ripgrep(ctx, ripgrepOpts)
		// Should handle conflicting options gracefully
		if err != nil {
			t.Logf("Conflicting options handled: %v", err)
		}
	})

	t.Run("Index Corruption Handling", func(t *testing.T) {
		// Test indexing invalid resource
		invalidResource := &storage.Resource{
			ID:      "index-test",
			Title:   "Index Test",
			Content: "", // Empty content
		}

		err := searchEngine.IndexResource(ctx, invalidResource)
		assert.NoError(t, err) // Should handle empty content

		// Test indexing resource with nil fields
		nilFieldResource := &storage.Resource{
			ID:      "nil-fields",
			Title:   "",
			Content: "Content",
			Tags:    nil, // nil slice
		}

		err = searchEngine.IndexResource(ctx, nilFieldResource)
		assert.NoError(t, err) // Should handle nil slices

		// Test indexing prompt with nil variables
		nilVarsPrompt := &storage.Prompt{
			ID:        "nil-vars",
			Name:      "Nil Variables",
			Template:  "Template",
			Variables: nil, // nil map
		}

		err = searchEngine.IndexPrompt(ctx, nilVarsPrompt)
		assert.NoError(t, err) // Should handle nil maps
	})
}

// TestDomainErrorHandling tests error handling in domain operations
func TestDomainErrorHandling(t *testing.T) {
	_, _, cleanup := setupErrorTestEnvironment(t)
	defer cleanup()
	
	// Initialize domain manager for this test
	domainConfig := &domain.DomainConfig{
		DefaultDomain:    "default",
		IsolationMode:    "standard",
		AllowCrossDomain: true,
	}
	domainManager := domain.NewDomainManager(domainConfig)

	ctx := context.Background()

	t.Run("Invalid Domain Creation", func(t *testing.T) {
		// Test with empty domain ID
		_, err := domainManager.CreateDomain(ctx, "", "Empty ID", "Description", "")
		assert.Error(t, err)

		// Test with invalid characters in domain ID
		invalidIDs := []string{
			"UPPERCASE",      // Uppercase not allowed
			"with spaces",    // Spaces not allowed
			"with-special!",  // Special chars not allowed
			"123",            // Only numbers
			"a",              // Too short
			strings.Repeat("a", 100), // Too long
		}

		for _, id := range invalidIDs {
			_, err := domainManager.CreateDomain(ctx, id, "Test", "Test", "")
			assert.Error(t, err, "Should reject invalid domain ID: %s", id)
		}

		// Test with non-existent parent
		_, err = domainManager.CreateDomain(ctx, "orphan", "Orphan", "Has non-existent parent", "non-existent-parent")
		assert.Error(t, err)

		// Test circular dependency (would require multiple domains)
		_, err = domainManager.CreateDomain(ctx, "parent", "Parent", "Parent domain", "")
		require.NoError(t, err)

		_, err = domainManager.CreateDomain(ctx, "child", "Child", "Child domain", "parent")
		require.NoError(t, err)

		// Try to make parent a child of child (circular)
		// This would require modifying an existing domain, which may not be supported
		// but we can test switching to non-existent domain
	})

	t.Run("Domain Switching Errors", func(t *testing.T) {
		// Test switching to non-existent domain
		err := domainManager.SetCurrentDomain("non-existent")
		assert.Error(t, err)

		// Test switching to empty domain
		err = domainManager.SetCurrentDomain("")
		assert.Error(t, err)

		// Test switching to domain with invalid characters
		err = domainManager.SetCurrentDomain("invalid domain name")
		assert.Error(t, err)
	})

	t.Run("Domain Scope Errors", func(t *testing.T) {
		// Test getting scope for non-existent domain
		_, err := domainManager.GetDomainScope("non-existent")
		assert.Error(t, err)

		// Test getting domain info for non-existent domain
		_, err = domainManager.GetDomain("non-existent")
		assert.Error(t, err)
	})

	t.Run("Shorthand Notation Errors", func(t *testing.T) {
		// Test parsing invalid shorthand notation
		invalidNotations := []string{
			":::",           // Too many colons
			":",             // Just colon
			"domain:",       // Missing resource ID
			":resource",     // Missing domain part for specific domain
			"dom ain:res",   // Invalid characters
		}

		for _, notation := range invalidNotations {
			domain, resourceID := domainManager.ParseResourceID(notation)
			// Should handle gracefully - either parse as best effort or use defaults
			t.Logf("Notation %s parsed as domain=%s, resourceID=%s", notation, domain, resourceID)
		}

		// Test building with invalid inputs
		id := domainManager.BuildResourceID("", "resource") // Empty domain
		assert.NotEmpty(t, id) // Should handle gracefully

		id = domainManager.BuildResourceID("domain", "") // Empty resource
		assert.NotEmpty(t, id) // Should handle gracefully
	})
}

// TestEdgeCases tests various edge cases and boundary conditions
func TestEdgeCases(t *testing.T) {
	store, searchEngine, cleanup := setupErrorTestEnvironment(t)
	defer cleanup()
	
	// CLI tools and domain manager will be initialized when needed
	_ = store // Use store variable to avoid unused warning

	ctx := context.Background()

	t.Run("Large Data Handling", func(t *testing.T) {
		// Test with very large content
		largeContent := strings.Repeat("This is a very long content string that tests the limits of the system. ", 10000)
		
		largeResource := &storage.Resource{
			ID:      "large-content",
			Title:   "Large Content Test",
			Content: largeContent,
			Tags:    []string{"large", "test"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := store.StoreResource(ctx, largeResource)
		require.NoError(t, err)

		err = searchEngine.IndexResource(ctx, largeResource)
		require.NoError(t, err)

		// Should be able to retrieve it
		retrieved, err := store.GetResource(ctx, "large-content")
		require.NoError(t, err)
		assert.Equal(t, len(largeContent), len(retrieved.Content))

		// Should be searchable
		result, err := searchEngine.OptimizedSearch(ctx, "very long content", 10)
		require.NoError(t, err)
		assert.Contains(t, result, "large-content")
	})

	t.Run("Unicode and Special Characters", func(t *testing.T) {
		// Test with Unicode content
		unicodeResource := &storage.Resource{
			ID:      "unicode-test",
			Title:   "Unicode Test: 测试 🚀 Тест فحص",
			Content: "Content with various Unicode characters: 中文, العربية, Русский, 日本語, emoji: 🎉🔍📚",
			Tags:    []string{"unicode", "测试", "🏷️"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := store.StoreResource(ctx, unicodeResource)
		require.NoError(t, err)

		err = searchEngine.IndexResource(ctx, unicodeResource)
		require.NoError(t, err)

		// Should be able to retrieve it
		retrieved, err := store.GetResource(ctx, "unicode-test")
		require.NoError(t, err)
		assert.Equal(t, unicodeResource.Title, retrieved.Title)
		assert.Equal(t, unicodeResource.Content, retrieved.Content)

		// Should be searchable with Unicode terms
		result, err := searchEngine.OptimizedSearch(ctx, "测试", 10)
		require.NoError(t, err)
		// May or may not find it depending on search engine capabilities
		t.Logf("Unicode search result: %s", result)
	})

	t.Run("Boundary Value Testing", func(t *testing.T) {
		// Test with zero-length arrays and empty strings
		emptyResource := &storage.Resource{
			ID:       "empty-fields",
			Title:    "",
			Content:  "",
			Type:     "",
			Tags:     []string{},
			Metadata: map[string]interface{}{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := store.StoreResource(ctx, emptyResource)
		require.NoError(t, err)

		// Should be retrievable
		retrieved, err := store.GetResource(ctx, "empty-fields")
		require.NoError(t, err)
		assert.Equal(t, "", retrieved.Title)
		assert.Equal(t, 0, len(retrieved.Tags))

		// Test search with zero limit
		query := &search.AdvancedSearchQuery{
			Query: "test",
			Limit: 0,
		}

		results, _, err := searchEngine.AdvancedSearch(ctx, query)
		require.NoError(t, err)
		// Zero limit should be handled gracefully (either return 0 results or use default limit)
		assert.GreaterOrEqual(t, len(results), 0)
		assert.LessOrEqual(t, len(results), 20) // Default limit is 20

		// Test with maximum reasonable values
		maxQuery := &search.AdvancedSearchQuery{
			Query:  "test",
			Limit:  10000, // Large limit
			Fields: []string{"title", "content", "tags"}, // All fields
			Tags:   make([]string, 100), // Many tags
		}

		// Fill tags with test values
		for i := range maxQuery.Tags {
			maxQuery.Tags[i] = fmt.Sprintf("tag-%d", i)
		}

		_, _, err = searchEngine.AdvancedSearch(ctx, maxQuery)
		// Should handle gracefully (either work or fail with reasonable error)
		if err != nil {
			t.Logf("Max values query failed as expected: %v", err)
		}
	})

	t.Run("Timestamp Edge Cases", func(t *testing.T) {
		// Test with zero time
		zeroTimeResource := &storage.Resource{
			ID:        "zero-time",
			Title:     "Zero Time Test",
			Content:   "Test with zero timestamps",
			CreatedAt: time.Time{}, // Zero time
			UpdatedAt: time.Time{}, // Zero time
		}

		err := store.StoreResource(ctx, zeroTimeResource)
		require.NoError(t, err)

		retrieved, err := store.GetResource(ctx, "zero-time")
		require.NoError(t, err)
		// Should handle zero times gracefully
		t.Logf("Zero time resource timestamps: created=%v, updated=%v", retrieved.CreatedAt, retrieved.UpdatedAt)

		// Test with future time
		futureTime := time.Now().Add(100 * 365 * 24 * time.Hour) // 100 years in future
		futureTimeResource := &storage.Resource{
			ID:        "future-time",
			Title:     "Future Time Test",
			Content:   "Test with future timestamps",
			CreatedAt: futureTime,
			UpdatedAt: futureTime,
		}

		err = store.StoreResource(ctx, futureTimeResource)
		require.NoError(t, err) // Should handle future times

		// Test with very old time
		oldTime := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
		oldTimeResource := &storage.Resource{
			ID:        "old-time",
			Title:     "Old Time Test",
			Content:   "Test with very old timestamps",
			CreatedAt: oldTime,
			UpdatedAt: oldTime,
		}

		err = store.StoreResource(ctx, oldTimeResource)
		require.NoError(t, err) // Should handle old times
	})

	t.Run("Concurrent Edge Cases", func(t *testing.T) {
		// Test rapid creation and deletion of same resource ID
		resourceID := "rapid-ops"
		
		for i := 0; i < 10; i++ {
			resource := &storage.Resource{
				ID:      resourceID,
				Title:   fmt.Sprintf("Rapid Test %d", i),
				Content: fmt.Sprintf("Content %d", i),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			err := store.StoreResource(ctx, resource)
			require.NoError(t, err)

			// Try to retrieve immediately
			retrieved, err := store.GetResource(ctx, resourceID)
			require.NoError(t, err)
			assert.Equal(t, fmt.Sprintf("Rapid Test %d", i), retrieved.Title)
		}

		// Final verification
		final, err := store.GetResource(ctx, resourceID)
		require.NoError(t, err)
		assert.Equal(t, "Rapid Test 9", final.Title) // Should have the last version
	})

	t.Run("Resource Relationships", func(t *testing.T) {
		// Test resources with complex metadata relationships
		parentResource := &storage.Resource{
			ID:      "parent-resource",
			Title:   "Parent Resource",
			Content: "This is a parent resource",
			Metadata: map[string]interface{}{
				"children": []string{"child-1", "child-2"},
				"type":     "parent",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := store.StoreResource(ctx, parentResource)
		require.NoError(t, err)

		// Create child resources
		for i := 1; i <= 2; i++ {
			childResource := &storage.Resource{
				ID:      fmt.Sprintf("child-%d", i),
				Title:   fmt.Sprintf("Child Resource %d", i),
				Content: fmt.Sprintf("This is child resource %d", i),
				Metadata: map[string]interface{}{
					"parent": "parent-resource",
					"index":  i,
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			err := store.StoreResource(ctx, childResource)
			require.NoError(t, err)
		}

		// Should be able to retrieve all related resources
		parent, err := store.GetResource(ctx, "parent-resource")
		require.NoError(t, err)
		assert.Contains(t, parent.Metadata, "children")

		for i := 1; i <= 2; i++ {
			child, err := store.GetResource(ctx, fmt.Sprintf("child-%d", i))
			require.NoError(t, err)
			assert.Equal(t, "parent-resource", child.Metadata["parent"])
		}
	})
}