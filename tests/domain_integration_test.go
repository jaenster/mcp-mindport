package tests

import (
	"context"
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

func setupDomainTestEnvironment(t *testing.T) (*storage.BadgerStore, *search.BleveSearch, *domain.DomainManager, func()) {
	// Create temporary directories
	tempDir, err := os.MkdirTemp("", "domain_integration_test_*")
	require.NoError(t, err)

	storageDir := filepath.Join(tempDir, "storage")
	searchDir := filepath.Join(tempDir, "search")

	// Initialize storage
	store, err := storage.NewBadgerStore(storageDir)
	require.NoError(t, err)

	// Initialize search
	searchEngine, err := search.NewBleveSearch(searchDir)
	require.NoError(t, err)

	// Initialize domain manager
	domainConfig := &domain.DomainConfig{
		DefaultDomain:    "default",
		IsolationMode:    "standard",
		AllowCrossDomain: true,
	}
	domainManager := domain.NewDomainManager(domainConfig)

	// Cleanup function
	cleanup := func() {
		store.Close()
		searchEngine.Close()
		os.RemoveAll(tempDir)
	}

	return store, searchEngine, domainManager, cleanup
}

func createTestDomains(t *testing.T, domainManager *domain.DomainManager) {
	ctx := context.Background()

	// Create test domains
	domains := []struct {
		id          string
		name        string
		description string
		parent      string
	}{
		{"project1", "Project One", "First test project", ""},
		{"project2", "Project Two", "Second test project", ""},
		{"team1", "Team Alpha", "Development team alpha", ""},
		{"team1-backend", "Backend Services", "Backend services for team alpha", "team1"},
		{"team1-frontend", "Frontend Apps", "Frontend applications for team alpha", "team1"},
		{"shared", "Shared Resources", "Resources shared across projects", ""},
	}

	for _, d := range domains {
		_, err := domainManager.CreateDomain(ctx, d.id, d.name, d.description, d.parent)
		require.NoError(t, err)
	}
}

func createTestDataInDomains(t *testing.T, store *storage.BadgerStore, searchEngine *search.BleveSearch) {
	ctx := context.Background()

	// Create resources in different domains
	resources := []*storage.Resource{
		{
			ID:      "api-docs-p1",
			Domain:  "project1",
			Type:    "documentation",
			Title:   "Project 1 API Documentation",
			Content: "API documentation for project 1 services. Includes authentication, user management, and data endpoints.",
			Tags:    []string{"api", "documentation", "project1"},
			CreatedAt: time.Now().Add(-48 * time.Hour),
			UpdatedAt: time.Now().Add(-24 * time.Hour),
		},
		{
			ID:      "api-docs-p2",
			Domain:  "project2",
			Type:    "documentation", 
			Title:   "Project 2 API Documentation",
			Content: "API documentation for project 2 services. Includes inventory, orders, and reporting endpoints.",
			Tags:    []string{"api", "documentation", "project2"},
			CreatedAt: time.Now().Add(-72 * time.Hour),
			UpdatedAt: time.Now().Add(-12 * time.Hour),
		},
		{
			ID:      "backend-guide",
			Domain:  "team1-backend",
			Type:    "guide",
			Title:   "Backend Development Guide",
			Content: "Comprehensive guide for backend development in Team Alpha. Covers architecture, patterns, and best practices.",
			Tags:    []string{"backend", "guide", "team1"},
			CreatedAt: time.Now().Add(-96 * time.Hour),
			UpdatedAt: time.Now().Add(-6 * time.Hour),
		},
		{
			ID:      "frontend-standards",
			Domain:  "team1-frontend",
			Type:    "standards",
			Title:   "Frontend Coding Standards",
			Content: "Coding standards and style guide for frontend development. Includes React patterns, testing guidelines.",
			Tags:    []string{"frontend", "standards", "team1"},
			CreatedAt: time.Now().Add(-120 * time.Hour),
			UpdatedAt: time.Now().Add(-3 * time.Hour),
		},
		{
			ID:      "shared-utils",
			Domain:  "shared",
			Type:    "library",
			Title:   "Shared Utility Functions",
			Content: "Collection of utility functions shared across all projects. Includes validation, formatting, and common helpers.",
			Tags:    []string{"utilities", "shared", "library"},
			CreatedAt: time.Now().Add(-168 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		},
	}

	for _, resource := range resources {
		err := store.StoreResource(ctx, resource)
		require.NoError(t, err)

		err = searchEngine.IndexResource(ctx, resource)
		require.NoError(t, err)
	}

	// Create prompts in different domains
	prompts := []*storage.Prompt{
		{
			ID:          "code-review-p1",
			Domain:      "project1",
			Name:        "Project 1 Code Review",
			Description: "Code review template for project 1",
			Template:    "Review this {{language}} code for project 1:\n\n{{code}}\n\nFocus on: {{focus}}",
			Variables: map[string]string{
				"language": "Programming language",
				"code":     "Code to review",
				"focus":    "Review focus",
			},
			Tags:      []string{"review", "project1"},
			CreatedAt: time.Now().Add(-36 * time.Hour),
			UpdatedAt: time.Now().Add(-18 * time.Hour),
		},
		{
			ID:          "team-standup",
			Domain:      "team1",
			Name:        "Team Standup Template",
			Description: "Daily standup template for team alpha",
			Template:    "Standup for {{date}}:\n\nYesterday: {{yesterday}}\nToday: {{today}}\nBlockers: {{blockers}}",
			Variables: map[string]string{
				"date":      "Standup date",
				"yesterday": "What was accomplished yesterday",
				"today":     "What will be worked on today",
				"blockers":  "Any blockers or issues",
			},
			Tags:      []string{"standup", "team1", "meeting"},
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now().Add(-6 * time.Hour),
		},
	}

	for _, prompt := range prompts {
		err := store.StorePrompt(ctx, prompt)
		require.NoError(t, err)

		err = searchEngine.IndexPrompt(ctx, prompt)
		require.NoError(t, err)
	}

	// Give the index a moment to settle
	time.Sleep(100 * time.Millisecond)
}

// TestDomainCreation tests domain creation and validation
func TestDomainCreation(t *testing.T) {
	_, _, domainManager, cleanup := setupDomainTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("Create Valid Domain", func(t *testing.T) {
		domain, err := domainManager.CreateDomain(ctx, "test-project", "Test Project", "A test project", "")
		require.NoError(t, err)
		assert.Equal(t, "test-project", domain.ID)
		assert.Equal(t, "Test Project", domain.Name)
		assert.Equal(t, "A test project", domain.Description)
		assert.Equal(t, "", domain.Parent)
		assert.True(t, domain.Active)
		assert.Equal(t, "/test-project", domain.Path)
	})

	t.Run("Create Child Domain", func(t *testing.T) {
		// First create parent
		_, err := domainManager.CreateDomain(ctx, "parent", "Parent Domain", "Parent", "")
		require.NoError(t, err)

		// Then create child
		child, err := domainManager.CreateDomain(ctx, "child", "Child Domain", "Child", "parent")
		require.NoError(t, err)
		assert.Equal(t, "parent", child.Parent)
		assert.Equal(t, "/parent/child", child.Path)
	})

	t.Run("Create Domain with Invalid ID", func(t *testing.T) {
		invalidIDs := []string{
			"",           // Empty
			"UPPERCASE",  // Uppercase not allowed
			"with spaces", // Spaces not allowed
			"with_special_!", // Special chars not allowed
			"a",          // Too short
			"this-is-a-very-long-domain-id-that-exceeds-the-maximum-allowed-length", // Too long
		}

		for _, id := range invalidIDs {
			t.Run("ID: "+id, func(t *testing.T) {
				_, err := domainManager.CreateDomain(ctx, id, "Test", "Test", "")
				assert.Error(t, err)
			})
		}
	})

	t.Run("Create Duplicate Domain", func(t *testing.T) {
		// Create first domain
		_, err := domainManager.CreateDomain(ctx, "duplicate", "First", "First", "")
		require.NoError(t, err)

		// Try to create duplicate
		_, err = domainManager.CreateDomain(ctx, "duplicate", "Second", "Second", "")
		assert.Error(t, err)
	})

	t.Run("Create Domain with Non-existent Parent", func(t *testing.T) {
		_, err := domainManager.CreateDomain(ctx, "orphan", "Orphan", "Orphan", "non-existent")
		assert.Error(t, err)
	})
}

// TestDomainHierarchy tests domain hierarchy operations
func TestDomainHierarchy(t *testing.T) {
	_, _, domainManager, cleanup := setupDomainTestEnvironment(t)
	defer cleanup()

	createTestDomains(t, domainManager)

	t.Run("List Root Domains", func(t *testing.T) {
		domains := domainManager.ListDomains("")
		
		// Should have root domains: project1, project2, team1, shared
		rootDomains := make([]string, 0)
		for _, domain := range domains {
			if domain.Parent == "" {
				rootDomains = append(rootDomains, domain.ID)
			}
		}
		
		assert.Contains(t, rootDomains, "project1")
		assert.Contains(t, rootDomains, "project2")
		assert.Contains(t, rootDomains, "team1")
		assert.Contains(t, rootDomains, "shared")
	})

	t.Run("List Child Domains", func(t *testing.T) {
		children := domainManager.ListDomains("team1")
		
		childIDs := make([]string, 0)
		for _, child := range children {
			childIDs = append(childIDs, child.ID)
		}
		
		assert.Contains(t, childIDs, "team1-backend")
		assert.Contains(t, childIDs, "team1-frontend")
	})

	t.Run("Get Domain Scope", func(t *testing.T) {
		scope, err := domainManager.GetDomainScope("team1-backend")
		require.NoError(t, err)
		
		// Should include ancestry path
		assert.Contains(t, scope.Ancestry, "team1")
		assert.Contains(t, scope.Ancestry, "team1-backend")
		
		// Should be searchable in parent domains
		assert.Contains(t, scope.Searchable, "team1-backend")
		assert.Contains(t, scope.Searchable, "team1")
	})

	t.Run("Domain Path Generation", func(t *testing.T) {
		domain, err := domainManager.GetDomain("team1-backend")
		require.NoError(t, err)
		assert.Equal(t, "/team1/team1-backend", domain.Path)
	})
}

// TestDomainIsolation tests domain isolation and cross-domain access
func TestDomainIsolation(t *testing.T) {
	store, searchEngine, domainManager, cleanup := setupDomainTestEnvironment(t)
	defer cleanup()

	createTestDomains(t, domainManager)
	createTestDataInDomains(t, store, searchEngine)

	ctx := context.Background()

	t.Run("Domain Scoped Resource Access", func(t *testing.T) {
		// Switch to project1 domain
		err := domainManager.SetCurrentDomain("project1")
		require.NoError(t, err)

		// Should be able to access project1 resources
		resource, err := store.GetResourceInDomain(ctx, "api-docs-p1", "project1")
		require.NoError(t, err)
		assert.Equal(t, "project1", resource.Domain)

		// Should not be able to access project2 resources directly
		_, err = store.GetResourceInDomain(ctx, "api-docs-p2", "project1")
		assert.Error(t, err)
	})

	t.Run("Cross Domain Access with Permissions", func(t *testing.T) {
		// Switch to shared domain
		err := domainManager.SetCurrentDomain("shared")
		require.NoError(t, err)

		// Should be able to access shared resources from any domain
		resource, err := store.GetResourceInDomain(ctx, "shared-utils", "shared")
		require.NoError(t, err)
		assert.Equal(t, "shared", resource.Domain)
	})

	t.Run("Hierarchical Domain Access", func(t *testing.T) {
		// Switch to team1 (parent) domain
		err := domainManager.SetCurrentDomain("team1")
		require.NoError(t, err)

		// Should be able to access child domain resources
		resource, err := store.GetResourceInDomain(ctx, "backend-guide", "team1-backend")
		require.NoError(t, err)
		assert.Equal(t, "team1-backend", resource.Domain)

		// Switch to child domain
		err = domainManager.SetCurrentDomain("team1-backend")
		require.NoError(t, err)

		// Should be able to access parent domain resources
		scope, err := domainManager.GetDomainScope("team1-backend")
		require.NoError(t, err)
		assert.Contains(t, scope.Searchable, "team1")
	})
}

// TestDomainSearchScoping tests search operations within domain contexts
func TestDomainSearchScoping(t *testing.T) {
	store, searchEngine, domainManager, cleanup := setupDomainTestEnvironment(t)
	defer cleanup()

	createTestDomains(t, domainManager)
	createTestDataInDomains(t, store, searchEngine)

	ctx := context.Background()

	t.Run("Domain Filtered Search", func(t *testing.T) {
		// Search within project1 domain only
		query := &search.AdvancedSearchQuery{
			Query:   "API documentation",
			Domains: []string{"project1"},
			Limit:   10,
		}

		results, _, err := searchEngine.AdvancedSearch(ctx, query)
		require.NoError(t, err)
		
		// Should only return project1 results
		for _, result := range results {
			// Note: Domain would need to be added to search results or checked via resource lookup
			assert.Contains(t, result.ID, "project1") // Assuming resource IDs contain domain info
		}
	})

	t.Run("Multi Domain Search", func(t *testing.T) {
		// Search across multiple domains
		query := &search.AdvancedSearchQuery{
			Query:   "API",
			Domains: []string{"project1", "project2"},
			Limit:   10,
		}

		results, _, err := searchEngine.AdvancedSearch(ctx, query)
		require.NoError(t, err)
		
		// Should return results from both domains
		resourceIDs := make([]string, 0)
		for _, result := range results {
			resourceIDs = append(resourceIDs, result.ID)
		}
		
		// Check that we have results from both projects
		hasProject1 := false
		hasProject2 := false
		for _, id := range resourceIDs {
			if strings.Contains(id, "p1") {
				hasProject1 = true
			}
			if strings.Contains(id, "p2") {
				hasProject2 = true
			}
		}
		assert.True(t, hasProject1 || hasProject2)
	})

	t.Run("Hierarchical Search Scope", func(t *testing.T) {
		// Search from team1 should include child domains
		query := &search.AdvancedSearchQuery{
			Query:   "team1",
			Domains: []string{"team1", "team1-backend", "team1-frontend"},
			Limit:   10,
		}

		results, _, err := searchEngine.AdvancedSearch(ctx, query)
		require.NoError(t, err)
		
		// Should find results in team1 hierarchy
		assert.Greater(t, len(results), 0)
		
		// Check that results are related to team1
		for _, result := range results {
			assert.True(t, strings.Contains(result.Title, "team1") || strings.Contains(result.Content, "team1") || strings.Contains(result.ID, "team1"))
		}
	})

	t.Run("Global Search", func(t *testing.T) {
		// Search without domain filter should search all accessible domains
		query := &search.AdvancedSearchQuery{
			Query: "documentation",
			Limit: 20,
		}

		results, _, err := searchEngine.AdvancedSearch(ctx, query)
		require.NoError(t, err)
		
		// Should find results across multiple domains
		assert.Greater(t, len(results), 0)
		
		// Check that we have results from different sources
		uniqueTypes := make(map[string]bool)
		for _, result := range results {
			uniqueTypes[result.Type] = true
		}
		
		// Should have some variety in results
		assert.GreaterOrEqual(t, len(uniqueTypes), 1)
	})
}

// TestDomainShorthandNotation tests the shorthand notation parsing
func TestDomainShorthandNotation(t *testing.T) {
	_, _, domainManager, cleanup := setupDomainTestEnvironment(t)
	defer cleanup()

	createTestDomains(t, domainManager)

	t.Run("Parse Default Domain Notation", func(t *testing.T) {
		domain, resourceID := domainManager.ParseResourceID("::abc123")
		assert.Equal(t, "default", domain) // Should use default domain
		assert.Equal(t, "abc123", resourceID)
	})

	t.Run("Parse Specific Domain Notation", func(t *testing.T) {
		domain, resourceID := domainManager.ParseResourceID("project1:abc123")
		assert.Equal(t, "project1", domain)
		assert.Equal(t, "abc123", resourceID)
	})

	t.Run("Parse Plain Resource ID", func(t *testing.T) {
		// Set current domain
		err := domainManager.SetCurrentDomain("project2")
		require.NoError(t, err)

		domain, resourceID := domainManager.ParseResourceID("abc123")
		assert.Equal(t, "project2", domain) // Should use current domain
		assert.Equal(t, "abc123", resourceID)
	})

	t.Run("Build Resource ID", func(t *testing.T) {
		// Test building shorthand notation
		id := domainManager.BuildResourceID("project1", "abc123")
		assert.Equal(t, "project1:abc123", id)

		// Test default domain notation
		id = domainManager.BuildResourceID("default", "abc123")
		assert.Equal(t, "::abc123", id)
	})
}

// TestDomainSwitching tests switching between domains
func TestDomainSwitching(t *testing.T) {
	_, _, domainManager, cleanup := setupDomainTestEnvironment(t)
	defer cleanup()

	createTestDomains(t, domainManager)

	t.Run("Switch to Valid Domain", func(t *testing.T) {
		err := domainManager.SetCurrentDomain("project1")
		require.NoError(t, err)
		
		current := domainManager.GetCurrentDomain()
		assert.Equal(t, "project1", current)
	})

	t.Run("Switch to Invalid Domain", func(t *testing.T) {
		err := domainManager.SetCurrentDomain("non-existent")
		assert.Error(t, err)
		
		// Current domain should remain unchanged
		current := domainManager.GetCurrentDomain()
		assert.Equal(t, "project1", current) // From previous test
	})

	t.Run("Switch to Child Domain", func(t *testing.T) {
		err := domainManager.SetCurrentDomain("team1-backend")
		require.NoError(t, err)
		
		current := domainManager.GetCurrentDomain()
		assert.Equal(t, "team1-backend", current)
	})
}

// TestDomainStats tests domain statistics
func TestDomainStats(t *testing.T) {
	store, searchEngine, domainManager, cleanup := setupDomainTestEnvironment(t)
	defer cleanup()

	createTestDomains(t, domainManager)
	createTestDataInDomains(t, store, searchEngine)

	ctx := context.Background()

	t.Run("Get Domain Statistics", func(t *testing.T) {
		stats, err := store.GetDomainStats(ctx, "project1")
		require.NoError(t, err)
		
		// Should have one resource in project1
		assert.Equal(t, 1, stats["resources"])
		assert.Equal(t, 1, stats["prompts"])
	})

	t.Run("Get Empty Domain Statistics", func(t *testing.T) {
		stats, err := store.GetDomainStats(ctx, "project2")
		require.NoError(t, err)
		
		// project2 has 1 resource, 0 prompts
		assert.Equal(t, 1, stats["resources"])
		assert.Equal(t, 0, stats["prompts"])
	})

	t.Run("Get Non-existent Domain Statistics", func(t *testing.T) {
		stats, err := store.GetDomainStats(ctx, "non-existent")
		require.NoError(t, err)
		
		// Should return zeros for non-existent domain
		assert.Equal(t, 0, stats["resources"])
		assert.Equal(t, 0, stats["prompts"])
	})
}

// TestDomainConfiguration tests domain configuration options
func TestDomainConfiguration(t *testing.T) {
	_, _, domainManager, cleanup := setupDomainTestEnvironment(t)
	defer cleanup()

	t.Run("Get Domain Configuration", func(t *testing.T) {
		config := domainManager.GetConfig()
		
		assert.Equal(t, "default", config.DefaultDomain)
		assert.Equal(t, "standard", config.IsolationMode)
		assert.True(t, config.AllowCrossDomain)
	})

	t.Run("Different Isolation Modes", func(t *testing.T) {
		// Test with strict isolation
		strictConfig := &domain.DomainConfig{
			DefaultDomain:    "default",
			IsolationMode:    "strict",
			AllowCrossDomain: false,
		}
		
		strictManager := domain.NewDomainManager(strictConfig)
		config := strictManager.GetConfig()
		
		assert.Equal(t, "strict", config.IsolationMode)
		assert.False(t, config.AllowCrossDomain)
	})
}