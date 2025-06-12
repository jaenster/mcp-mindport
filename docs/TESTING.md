# Testing Documentation

Comprehensive testing strategy and test suite documentation for MindPort MCP Server.

## Test Overview

MindPort includes a comprehensive test suite with **76 tests** covering all functionality:

- **Storage Layer Tests** (17 tests): SQLite operations, CRUD, domain management
- **Search Engine Tests** (30 tests): Fuzzy search, patterns, grep, tag filtering  
- **MCP Server Tests** (25 tests): All MCP tools, error handling, schemas
- **Integration Tests** (4 tests): End-to-end workflows, performance, multi-domain

## Running Tests

### Basic Test Commands

```bash
# Run all tests in watch mode
npm test

# Run tests once and exit
npm run test:run

# Run tests with UI interface
npm run test:ui

# Run tests with coverage report
npm run test:coverage
```

### Specific Test Suites

```bash
# Run specific test files
npx vitest tests/storage.test.ts
npx vitest tests/search.test.ts  
npx vitest tests/server.test.ts
npx vitest tests/integration.test.ts

# Run tests matching pattern
npx vitest --run --reporter=verbose tests/storage

# Run single test
npx vitest --run -t "should store a resource"
```

## Test Architecture

### Test Framework: Vitest

- **Fast**: Native ESM support, parallel execution
- **Modern**: TypeScript support, mocking capabilities
- **Compatible**: Jest-like API with better performance

### Test Structure

```
tests/
├── setup.ts              # Global test configuration
├── storage.test.ts        # SQLite storage layer tests
├── search.test.ts         # Fuse.js search engine tests  
├── server.test.ts         # MCP server and tools tests
└── integration.test.ts    # End-to-end workflow tests
```

### Configuration

**vitest.config.ts:**
```typescript
export default defineConfig({
  test: {
    globals: true,
    environment: 'node',
    testTimeout: 10000,
    hookTimeout: 10000,
    teardownTimeout: 5000,
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html']
    },
    setupFiles: ['./tests/setup.ts']
  }
});
```

## Test Categories

### 1. Storage Layer Tests (17 tests)

Tests SQLite database operations and domain management.

**Key Test Areas:**
- Domain creation and management
- Resource CRUD operations  
- Prompt template storage
- Data validation and constraints
- Concurrent access handling

**Example Test:**
```typescript
describe('Resource Management', () => {
  it('should store a resource', async () => {
    await storage.storeResource({
      id: 'test-resource',
      name: 'Test Resource',
      content: 'Test content',
      tags: ['test'],
      domain: 'default'
    });
    
    const retrieved = await storage.getResource('test-resource');
    expect(retrieved?.name).toBe('Test Resource');
  });
});
```

### 2. Search Engine Tests (30 tests)

Tests Fuse.js search functionality and pattern matching.

**Key Test Areas:**
- Fuzzy search with scoring
- Regex pattern matching
- Grep-like functionality
- Tag-based filtering
- Search index management

**Example Test:**
```typescript
describe('Fuzzy Search', () => {
  it('should find resources by content', () => {
    const results = search.search('JavaScript programming');
    
    expect(results.length).toBeGreaterThan(0);
    expect(results[0].score).toBeGreaterThan(0.5);
    expect(results[0].matches).toContain('JavaScript');
  });
});
```

### 3. MCP Server Tests (25 tests)

Tests MCP protocol implementation and all tools.

**Key Test Areas:**
- Tool schema validation
- Request/response handling
- Error handling and edge cases
- Domain switching and isolation
- Tool integration

**Example Test:**
```typescript
describe('MCP Tools', () => {
  it('should handle store_resource tool', async () => {
    const result = await server.handleStoreResource({
      id: 'test-resource',
      name: 'Test Resource', 
      content: 'Test content'
    });
    
    expect(result.content[0].text).toContain('stored successfully');
  });
});
```

### 4. Integration Tests (4 tests)

Tests complete workflows and system integration.

**Key Test Areas:**
- End-to-end resource lifecycle
- Multi-domain workflows  
- Performance with large datasets
- Complex search scenarios

**Example Test:**
```typescript
describe('End-to-End Workflow', () => {
  it('should handle complete resource lifecycle', async () => {
    // Create domain
    await storage.createDomain('test-project');
    
    // Store multiple resources
    // Update search index  
    // Perform various searches
    // Verify domain isolation
    
    expect(/* comprehensive assertions */);
  });
});
```

## Test Data and Fixtures

### Sample Resources

Tests use realistic sample data:

```typescript
const sampleResources = [
  {
    id: 'js-tutorial',
    name: 'JavaScript Tutorial',
    description: 'Learn JavaScript fundamentals',
    content: 'JavaScript is a programming language...',
    tags: ['javascript', 'tutorial', 'programming'],
    domain: 'default'
  },
  // ... more test data
];
```

### Test Database Isolation

Each test gets its own SQLite database:

```typescript
beforeEach(async () => {
  testDbPath = path.join(os.tmpdir(), `test-${Date.now()}-${Math.random()}.db`);
  storage = new SQLiteStorage(testDbPath);
  await storage.initialize();
});

afterEach(async () => {
  await storage.close();
  await fs.unlink(testDbPath);
});
```

## Performance Testing

### Performance Benchmarks

Integration tests include performance validation:

```typescript
it('should handle large datasets efficiently', async () => {
  // Store 100 resources
  const startTime = Date.now();
  for (let i = 0; i < 100; i++) {
    await storage.storeResource(/* resource */);
  }
  const storeTime = Date.now() - startTime;
  
  // Performance assertions
  expect(storeTime).toBeLessThan(10000); // < 10 seconds
  
  // Test search performance
  const searchStart = Date.now();
  const results = search.search('test query');
  const searchTime = Date.now() - searchStart;
  
  expect(searchTime).toBeLessThan(100); // < 100ms
});
```

### Expected Performance

- **Storage**: < 10s for 100 resources
- **Search**: < 100ms for queries
- **Indexing**: < 1s for index updates
- **Grep**: < 100ms for pattern matching

## Mocking Strategy

### MCP SDK Mocking

The MCP SDK is mocked to avoid stdio transport issues during testing:

```typescript
vi.mock('@modelcontextprotocol/sdk/server/index.js', () => ({
  Server: vi.fn().mockImplementation(() => ({
    setRequestHandler: vi.fn(),
    connect: vi.fn(),
    close: vi.fn(),
  })),
}));
```

### Database Mocking

Real SQLite databases are used (not mocked) to ensure integration correctness, but each test gets an isolated temporary database.

## Coverage Requirements

Target coverage levels:
- **Statements**: > 90%
- **Functions**: > 95%  
- **Lines**: > 90%
- **Branches**: > 85%

```bash
# Generate coverage report
npm run test:coverage

# View HTML coverage report
open coverage/index.html
```

## Debugging Tests

### Debug Individual Tests

```bash
# Debug specific test with verbose output
npx vitest --run --reporter=verbose -t "should store a resource"

# Run tests in debug mode
npx vitest --inspect-brk tests/storage.test.ts
```

### Test Debugging Tips

1. **Use `console.log`** strategically in tests
2. **Check database state** with temporary logging
3. **Verify test isolation** by running tests individually
4. **Use `expect.assertions()`** to ensure async tests complete

### Common Test Issues

**SQLite READONLY errors:**
```typescript
// Ensure unique database paths
testDbPath = path.join(os.tmpdir(), `test-${Date.now()}-${Math.random()}.db`);
```

**Async test timeouts:**
```typescript
// Increase timeout for slow operations
it('slow test', async () => {
  // test implementation
}, 30000); // 30 second timeout
```

**Search index synchronization:**
```typescript
// Always refresh search index before testing
await refreshSearchIndex();
const results = search.search(query);
```

## Test Automation

### CI/CD Integration

Tests are designed to run in CI environments:

```yaml
# Example GitHub Actions
- name: Run tests
  run: npm run test:run
  
- name: Generate coverage
  run: npm run test:coverage
  
- name: Upload coverage
  uses: codecov/codecov-action@v3
```

### Pre-commit Hooks

Recommended pre-commit setup:

```json
{
  "husky": {
    "hooks": {
      "pre-commit": "npm run test:run"
    }
  }
}
```

## Writing New Tests

### Test Structure Template

```typescript
import { describe, it, expect, beforeEach, afterEach } from 'vitest';

describe('Feature Name', () => {
  let testInstance: YourClass;
  
  beforeEach(async () => {
    // Setup test environment
    testInstance = new YourClass();
    await testInstance.initialize();
  });
  
  afterEach(async () => {
    // Cleanup test environment
    await testInstance.cleanup();
  });
  
  describe('Specific Functionality', () => {
    it('should handle expected behavior', async () => {
      // Arrange
      const input = { /* test data */ };
      
      // Act
      const result = await testInstance.method(input);
      
      // Assert
      expect(result).toBeDefined();
      expect(result.property).toBe(expectedValue);
    });
    
    it('should handle error conditions', async () => {
      // Test error scenarios
      await expect(
        testInstance.method(invalidInput)
      ).rejects.toThrow('Expected error message');
    });
  });
});
```

### Best Practices

1. **Test one thing per test** - Keep tests focused and atomic
2. **Use descriptive names** - Test names should explain the behavior
3. **Test both success and failure** - Cover happy path and edge cases
4. **Clean up resources** - Always close databases and clean up files
5. **Use realistic data** - Test with data similar to production use
6. **Verify side effects** - Ensure operations have expected side effects
7. **Test async operations** - Properly handle promises and async/await

This comprehensive testing strategy ensures MindPort is reliable, performant, and maintainable across all its functionality.