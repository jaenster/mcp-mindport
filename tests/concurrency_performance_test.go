package tests

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"mcp-mindport/internal/domain"
	"mcp-mindport/internal/search"
	"mcp-mindport/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupConcurrencyTestEnvironment(t *testing.T) (*storage.BadgerStore, *search.BleveSearch, *search.CLISearchTools, *domain.DomainManager, func()) {
	// Create temporary directories
	tempDir, err := os.MkdirTemp("", "concurrency_test_*")
	require.NoError(t, err)

	storageDir := filepath.Join(tempDir, "storage")
	searchDir := filepath.Join(tempDir, "search")

	// Initialize storage
	store, err := storage.NewBadgerStore(storageDir)
	require.NoError(t, err)

	// Initialize search
	searchEngine, err := search.NewBleveSearch(searchDir)
	require.NoError(t, err)

	// Initialize CLI tools
	cliTools := search.NewCLISearchTools(store, searchEngine)

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

	return store, searchEngine, cliTools, domainManager, cleanup
}

// TestConcurrentResourceOperations tests concurrent resource storage and retrieval
func TestConcurrentResourceOperations(t *testing.T) {
	store, searchEngine, _, _, cleanup := setupConcurrencyTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()
	numWorkers := runtime.NumCPU()
	resourcesPerWorker := 50

	t.Run("Concurrent Resource Storage", func(t *testing.T) {
		var wg sync.WaitGroup
		var errorMutex sync.Mutex
		var errors []error
		start := time.Now()

		// Start concurrent workers
		for workerID := 0; workerID < numWorkers; workerID++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				
				for i := 0; i < resourcesPerWorker; i++ {
					resourceID := fmt.Sprintf("worker-%d-resource-%d", id, i)
					resource := &storage.Resource{
						ID:      resourceID,
						Domain:  "default",
						Type:    "test",
						Title:   fmt.Sprintf("Worker %d Resource %d", id, i),
						Content: fmt.Sprintf("Content created by worker %d for resource %d. This includes some searchable content with unique identifier %s.", id, i, resourceID),
						Tags:    []string{"concurrent", "test", fmt.Sprintf("worker-%d", id)},
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}

					// Store resource
					if err := store.StoreResource(ctx, resource); err != nil {
						errorMutex.Lock()
						errors = append(errors, fmt.Errorf("worker %d failed to store resource %d: %v", id, i, err))
						errorMutex.Unlock()
						return
					}

					// Index resource
					if err := searchEngine.IndexResource(ctx, resource); err != nil {
						errorMutex.Lock()
						errors = append(errors, fmt.Errorf("worker %d failed to index resource %d: %v", id, i, err))
						errorMutex.Unlock()
						return
					}
				}
			}(workerID)
		}

		// Wait for all workers to complete
		wg.Wait()

		elapsed := time.Since(start)
		totalResources := numWorkers * resourcesPerWorker

		// Check for errors
		for _, err := range errors {
			t.Error(err)
		}

		t.Logf("Stored %d resources concurrently using %d workers in %v", totalResources, numWorkers, elapsed)
		assert.Less(t, elapsed, 30*time.Second) // Should complete in reasonable time

		// Verify all resources were stored
		time.Sleep(2 * time.Second) // Allow more time for database consistency

		for workerID := 0; workerID < numWorkers; workerID++ {
			for i := 0; i < min(resourcesPerWorker, 5); i++ { // Sample verification
				resourceID := fmt.Sprintf("worker-%d-resource-%d", workerID, i)
				retrieved, err := store.GetResource(ctx, resourceID)
				require.NoError(t, err, "Failed to retrieve resource %s", resourceID)
				assert.Equal(t, resourceID, retrieved.ID)
			}
		}
	})

	t.Run("Concurrent Resource Retrieval", func(t *testing.T) {
		var wg sync.WaitGroup
		var errorMutex sync.Mutex
		var errors []error
		retrievalCount := 100

		start := time.Now()

		// Start concurrent readers
		for workerID := 0; workerID < numWorkers; workerID++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				
				for i := 0; i < retrievalCount; i++ {
					// Randomly select a resource to retrieve
					targetWorker := rand.Intn(numWorkers)
					targetResource := rand.Intn(resourcesPerWorker)
					resourceID := fmt.Sprintf("worker-%d-resource-%d", targetWorker, targetResource)

					retrieved, err := store.GetResource(ctx, resourceID)
					if err != nil {
						errorMutex.Lock()
						errors = append(errors, fmt.Errorf("worker %d failed to retrieve %s: %v", id, resourceID, err))
						errorMutex.Unlock()
						continue
					}

					if retrieved.ID != resourceID {
						errorMutex.Lock()
						errors = append(errors, fmt.Errorf("worker %d got wrong resource: expected %s, got %s", id, resourceID, retrieved.ID))
						errorMutex.Unlock()
					}
				}
			}(workerID)
		}

		wg.Wait()

		elapsed := time.Since(start)
		totalRetrievals := numWorkers * retrievalCount

		// Check for errors
		errorCount := len(errors)
		for i, err := range errors {
			if i < 10 { // Only log first 10 errors
				t.Error(err)
			}
		}

		t.Logf("Performed %d concurrent retrievals using %d workers in %v", totalRetrievals, numWorkers, elapsed)
		assert.Less(t, float64(errorCount)/float64(totalRetrievals), 0.01) // Less than 1% error rate
	})
}

// TestConcurrentSearchOperations tests concurrent search operations
func TestConcurrentSearchOperations(t *testing.T) {
	store, searchEngine, cliTools, _, cleanup := setupConcurrencyTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	// First, create some test data
	numTestResources := 200
	for i := 0; i < numTestResources; i++ {
		resource := &storage.Resource{
			ID:      fmt.Sprintf("search-test-%d", i),
			Domain:  "default",
			Type:    "document",
			Title:   fmt.Sprintf("Search Test Document %d", i),
			Content: fmt.Sprintf("This is test document %d containing keywords like search, test, document, and topic-%d.", i, i%10),
			Tags:    []string{"search", "test", fmt.Sprintf("topic-%d", i%10)},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := store.StoreResource(ctx, resource)
		require.NoError(t, err)

		err = searchEngine.IndexResource(ctx, resource)
		require.NoError(t, err)
	}

	// Allow index to settle
	time.Sleep(2 * time.Second)

	numWorkers := runtime.NumCPU()
	searchesPerWorker := 50

	t.Run("Concurrent Basic Search", func(t *testing.T) {
		var wg sync.WaitGroup
		var errorMutex sync.Mutex
		var resultMutex sync.Mutex
		var errors []error
		var resultCounts []int

		start := time.Now()

		for workerID := 0; workerID < numWorkers; workerID++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				
				for i := 0; i < searchesPerWorker; i++ {
					// Vary the search terms
					searchTerm := fmt.Sprintf("topic-%d", (id+i)%10)
					
					result, err := searchEngine.OptimizedSearch(ctx, searchTerm, 10)
					if err != nil {
						errorMutex.Lock()
						errors = append(errors, fmt.Errorf("worker %d search %d failed: %v", id, i, err))
						errorMutex.Unlock()
						continue
					}

					// Count results (rough estimate)
					resultCount := len(result) / 100 // Rough heuristic
					resultMutex.Lock()
					resultCounts = append(resultCounts, resultCount)
					resultMutex.Unlock()
				}
			}(workerID)
		}

		wg.Wait()

		elapsed := time.Since(start)
		totalSearches := numWorkers * searchesPerWorker

		// Check for errors
		for _, err := range errors {
			t.Error(err)
		}

		// Verify results
		totalResults := 0
		for _, count := range resultCounts {
			totalResults += count
		}

		t.Logf("Performed %d concurrent searches using %d workers in %v", totalSearches, numWorkers, elapsed)
		assert.Greater(t, totalResults, 0)
		assert.Less(t, elapsed, 10*time.Second)
	})

	t.Run("Concurrent Advanced Search", func(t *testing.T) {
		var wg sync.WaitGroup
		var errorMutex sync.Mutex
		var errors []error
		start := time.Now()

		for workerID := 0; workerID < numWorkers; workerID++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				
				for i := 0; i < searchesPerWorker/2; i++ { // Fewer advanced searches
					query := &search.AdvancedSearchQuery{
						Query: fmt.Sprintf("document topic-%d", (id+i)%10),
						Mode:  "smart",
						Limit: 10,
					}

					results, _, err := searchEngine.AdvancedSearch(ctx, query)
					if err != nil {
						errorMutex.Lock()
						errors = append(errors, fmt.Errorf("worker %d advanced search %d failed: %v", id, i, err))
						errorMutex.Unlock()
						continue
					}

					assert.LessOrEqual(t, len(results), 10)
				}
			}(workerID)
		}

		wg.Wait()

		elapsed := time.Since(start)

		// Check for errors
		for _, err := range errors {
			t.Error(err)
		}

		t.Logf("Performed %d concurrent advanced searches using %d workers in %v", numWorkers*searchesPerWorker/2, numWorkers, elapsed)
	})

	t.Run("Concurrent CLI Tools", func(t *testing.T) {
		var wg sync.WaitGroup
		var errorMutex sync.Mutex
		var errors []error

		start := time.Now()

		for workerID := 0; workerID < numWorkers; workerID++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				
				// Test different CLI tools concurrently
				tools := []func() error{
					func() error {
						opts := &search.GrepOptions{Pattern: "test"}
						_, err := cliTools.Grep(ctx, opts)
						return err
					},
					func() error {
						opts := &search.FindOptions{Name: ".*Test.*"}
						_, err := cliTools.Find(ctx, opts)
						return err
					},
					func() error {
						opts := &search.RipgrepOptions{Pattern: "document"}
						_, err := cliTools.Ripgrep(ctx, opts)
						return err
					},
				}

				for i, tool := range tools {
					if err := tool(); err != nil {
						errorMutex.Lock()
						errors = append(errors, fmt.Errorf("worker %d tool %d failed: %v", id, i, err))
						errorMutex.Unlock()
					}
				}
			}(workerID)
		}

		wg.Wait()

		elapsed := time.Since(start)

		// Check for errors
		for _, err := range errors {
			t.Error(err)
		}

		t.Logf("Performed concurrent CLI tool operations using %d workers in %v", numWorkers, elapsed)
	})
}

// TestMixedConcurrentOperations tests mixed read/write operations
func TestMixedConcurrentOperations(t *testing.T) {
	store, searchEngine, _, _, cleanup := setupConcurrencyTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()
	numWriters := runtime.NumCPU() / 2
	numReaders := runtime.NumCPU()
	operationsPerWorker := 25

	// Pre-populate with some data
	for i := 0; i < 50; i++ {
		resource := &storage.Resource{
			ID:      fmt.Sprintf("mixed-test-%d", i),
			Domain:  "default",
			Type:    "mixed",
			Title:   fmt.Sprintf("Mixed Test Resource %d", i),
			Content: fmt.Sprintf("Initial content for mixed test resource %d", i),
			Tags:    []string{"mixed", "test"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := store.StoreResource(ctx, resource)
		require.NoError(t, err)

		err = searchEngine.IndexResource(ctx, resource)
		require.NoError(t, err)
	}

	time.Sleep(1 * time.Second)

	t.Run("Concurrent Read/Write Operations", func(t *testing.T) {
		var wg sync.WaitGroup
		var errorMutex sync.Mutex
		var errors []error
		
		start := time.Now()

		// Start writers
		for writerID := 0; writerID < numWriters; writerID++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				
				for i := 0; i < operationsPerWorker; i++ {
					resourceID := fmt.Sprintf("writer-%d-resource-%d", id, i)
					resource := &storage.Resource{
						ID:      resourceID,
						Domain:  "default",
						Type:    "concurrent",
						Title:   fmt.Sprintf("Concurrent Writer %d Resource %d", id, i),
						Content: fmt.Sprintf("Content written by writer %d at iteration %d", id, i),
						Tags:    []string{"concurrent", "writer", fmt.Sprintf("writer-%d", id)},
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}

					if err := store.StoreResource(ctx, resource); err != nil {
						errorMutex.Lock()
						errors = append(errors, fmt.Errorf("writer %d failed: %v", id, err))
						errorMutex.Unlock()
						return
					}

					if err := searchEngine.IndexResource(ctx, resource); err != nil {
						errorMutex.Lock()
						errors = append(errors, fmt.Errorf("writer %d index failed: %v", id, err))
						errorMutex.Unlock()
						return
					}

					// Occasional update to existing resource
					if i%5 == 0 && i > 0 {
						updateID := fmt.Sprintf("mixed-test-%d", rand.Intn(50))
						existing, err := store.GetResource(ctx, updateID)
						if err == nil {
							existing.Content = fmt.Sprintf("UPDATED by writer %d: %s", id, existing.Content)
							existing.UpdatedAt = time.Now()
							
							store.StoreResource(ctx, existing)
							searchEngine.IndexResource(ctx, existing)
						}
					}
				}
			}(writerID)
		}

		// Start readers
		for readerID := 0; readerID < numReaders; readerID++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				
				for i := 0; i < operationsPerWorker; i++ {
					// Mix of direct retrieval and search operations
					if i%2 == 0 {
						// Direct retrieval
						resourceID := fmt.Sprintf("mixed-test-%d", rand.Intn(50))
						_, err := store.GetResource(ctx, resourceID)
						if err != nil {
							// Don't fail on not found - might be updating
							continue
						}
					} else {
						// Search operation
						query := &search.AdvancedSearchQuery{
							Query: "test",
							Limit: 5,
						}
						_, _, err := searchEngine.AdvancedSearch(ctx, query)
						if err != nil {
							errorMutex.Lock()
							errors = append(errors, fmt.Errorf("reader %d search failed: %v", id, err))
							errorMutex.Unlock()
							return
						}
					}
				}
			}(readerID)
		}

		wg.Wait()

		elapsed := time.Since(start)

		// Check for errors
		errorCount := len(errors)
		for i, err := range errors {
			if i < 5 {
				t.Error(err)
			}
		}

		totalOperations := (numWriters + numReaders) * operationsPerWorker
		t.Logf("Performed %d mixed operations (%d writers, %d readers) in %v", totalOperations, numWriters, numReaders, elapsed)
		assert.Less(t, float64(errorCount)/float64(totalOperations), 0.05) // Less than 5% error rate
	})
}

// TestMemoryAndResourceUsage tests memory usage under concurrent load
func TestMemoryAndResourceUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	store, searchEngine, _, _, cleanup := setupConcurrencyTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("Memory Usage Under Load", func(t *testing.T) {
		var memStatsBefore, memStatsAfter runtime.MemStats
		
		// Force GC and get initial memory stats
		runtime.GC()
		runtime.ReadMemStats(&memStatsBefore)

		// Create significant load
		numResources := 1000
		for i := 0; i < numResources; i++ {
			resource := &storage.Resource{
				ID:      fmt.Sprintf("memory-test-%d", i),
				Domain:  "default",
				Type:    "memory",
				Title:   fmt.Sprintf("Memory Test Resource %d", i),
				Content: fmt.Sprintf("Content for memory test %d. %s", i, generateLargeContent(1024)), // 1KB content
				Tags:    []string{"memory", "test", fmt.Sprintf("batch-%d", i/100)},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			err := store.StoreResource(ctx, resource)
			require.NoError(t, err)

			err = searchEngine.IndexResource(ctx, resource)
			require.NoError(t, err)

			// Perform search every 100 resources
			if i%100 == 0 {
				query := &search.AdvancedSearchQuery{
					Query: "memory",
					Limit: 10,
				}
				_, _, err := searchEngine.AdvancedSearch(ctx, query)
				require.NoError(t, err)
			}
		}

		// Force GC and get final memory stats
		runtime.GC()
		runtime.ReadMemStats(&memStatsAfter)

		memoryIncrease := memStatsAfter.Alloc - memStatsBefore.Alloc
		t.Logf("Memory increase after storing %d resources: %d bytes (%.2f MB)", 
			numResources, memoryIncrease, float64(memoryIncrease)/(1024*1024))

		// Memory increase should be reasonable (less than 100MB for 1000 1KB resources)
		assert.Less(t, memoryIncrease, uint64(100*1024*1024))
	})
}

// TestDeadlockPrevention tests for potential deadlocks
func TestDeadlockPrevention(t *testing.T) {
	store, searchEngine, _, domainManager, cleanup := setupConcurrencyTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("No Deadlocks Under Concurrent Load", func(t *testing.T) {
		// Create test domains
		for i := 0; i < 5; i++ {
			domainID := fmt.Sprintf("deadlock-test-%d", i)
			_, err := domainManager.CreateDomain(ctx, domainID, fmt.Sprintf("Deadlock Test %d", i), "Test domain", "")
			require.NoError(t, err)
		}

		numWorkers := 10
		operationsPerWorker := 20
		timeout := 30 * time.Second

		done := make(chan bool, numWorkers)
		var wg sync.WaitGroup

		// Create a context with timeout to detect deadlocks
		testCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		for workerID := 0; workerID < numWorkers; workerID++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				defer func() { done <- true }()

				for i := 0; i < operationsPerWorker; i++ {
					select {
					case <-testCtx.Done():
						return // Timeout - potential deadlock
					default:
					}

					domain := fmt.Sprintf("deadlock-test-%d", i%5)
					resourceID := fmt.Sprintf("worker-%d-op-%d", id, i)

					// Mix of operations that could potentially deadlock
					switch i % 4 {
					case 0:
						// Store resource
						resource := &storage.Resource{
							ID:      resourceID,
							Domain:  domain,
							Type:    "deadlock",
							Title:   fmt.Sprintf("Deadlock Test %d", i),
							Content: "Test content",
							Tags:    []string{"deadlock", "test"},
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						}
						store.StoreResource(testCtx, resource)
						searchEngine.IndexResource(testCtx, resource)

					case 1:
						// Switch domains
						domainManager.SetCurrentDomain(domain)

					case 2:
						// Search
						query := &search.AdvancedSearchQuery{
							Query:   "deadlock",
							Domains: []string{domain},
							Limit:   5,
						}
						searchEngine.AdvancedSearch(testCtx, query)

					case 3:
						// Retrieve
						store.GetResource(testCtx, resourceID)
					}
				}
			}(workerID)
		}

		// Wait for completion or timeout
		completed := 0
		for i := 0; i < numWorkers; i++ {
			select {
			case <-done:
				completed++
			case <-time.After(timeout + 5*time.Second):
				t.Fatalf("Potential deadlock detected - only %d/%d workers completed", completed, numWorkers)
			}
		}

		wg.Wait()
		t.Logf("All %d workers completed successfully - no deadlocks detected", numWorkers)
	})
}

// TestPerformanceBenchmarks provides performance benchmarks
func TestPerformanceBenchmarks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance benchmarks in short mode")
	}

	store, searchEngine, _, _, cleanup := setupConcurrencyTestEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	// Prepare test data
	numResources := 500
	for i := 0; i < numResources; i++ {
		resource := &storage.Resource{
			ID:      fmt.Sprintf("perf-test-%d", i),
			Domain:  "default",
			Type:    "performance",
			Title:   fmt.Sprintf("Performance Test Resource %d", i),
			Content: fmt.Sprintf("Performance test content %d with keywords %s", i, generateSearchableContent(i)),
			Tags:    []string{"performance", "test", fmt.Sprintf("category-%d", i%10)},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := store.StoreResource(ctx, resource)
		require.NoError(t, err)

		err = searchEngine.IndexResource(ctx, resource)
		require.NoError(t, err)
	}

	time.Sleep(2 * time.Second) // Let index settle

	t.Run("Storage Performance", func(t *testing.T) {
		numOps := 100
		start := time.Now()

		for i := 0; i < numOps; i++ {
			resource := &storage.Resource{
				ID:      fmt.Sprintf("storage-perf-%d", i),
				Domain:  "default",
				Type:    "perf",
				Title:   fmt.Sprintf("Storage Performance %d", i),
				Content: "Performance test content",
				Tags:    []string{"perf"},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			err := store.StoreResource(ctx, resource)
			require.NoError(t, err)
		}

		elapsed := time.Since(start)
		opsPerSecond := float64(numOps) / elapsed.Seconds()
		t.Logf("Storage: %d operations in %v (%.2f ops/sec)", numOps, elapsed, opsPerSecond)
		
		assert.Greater(t, opsPerSecond, 50.0) // Should handle at least 50 ops/sec
	})

	t.Run("Search Performance", func(t *testing.T) {
		numOps := 100
		start := time.Now()

		for i := 0; i < numOps; i++ {
			query := &search.AdvancedSearchQuery{
				Query: fmt.Sprintf("test %d", i%10),
				Limit: 10,
			}

			_, _, err := searchEngine.AdvancedSearch(ctx, query)
			require.NoError(t, err)
		}

		elapsed := time.Since(start)
		opsPerSecond := float64(numOps) / elapsed.Seconds()
		t.Logf("Search: %d operations in %v (%.2f ops/sec)", numOps, elapsed, opsPerSecond)
		
		assert.Greater(t, opsPerSecond, 20.0) // Should handle at least 20 searches/sec
	})

	t.Run("Retrieval Performance", func(t *testing.T) {
		numOps := 200
		start := time.Now()

		for i := 0; i < numOps; i++ {
			resourceID := fmt.Sprintf("perf-test-%d", i%numResources)
			_, err := store.GetResource(ctx, resourceID)
			require.NoError(t, err)
		}

		elapsed := time.Since(start)
		opsPerSecond := float64(numOps) / elapsed.Seconds()
		t.Logf("Retrieval: %d operations in %v (%.2f ops/sec)", numOps, elapsed, opsPerSecond)
		
		assert.Greater(t, opsPerSecond, 100.0) // Should handle at least 100 retrievals/sec
	})
}

// Helper functions

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func generateLargeContent(size int) string {
	content := make([]byte, size)
	for i := range content {
		content[i] = byte('a' + (i % 26))
	}
	return string(content)
}

func generateSearchableContent(seed int) string {
	keywords := []string{"performance", "test", "benchmark", "concurrent", "search", "storage", "indexing", "retrieval"}
	return fmt.Sprintf("%s %s %s", keywords[seed%len(keywords)], keywords[(seed+1)%len(keywords)], keywords[(seed+2)%len(keywords)])
}