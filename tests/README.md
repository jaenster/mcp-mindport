# Integration Tests

This directory contains comprehensive integration tests for the MCP MindPort server.

## Test Structure

### Core Test Files

- **`mcp_integration_test.go`** - MCP protocol integration tests
  - Server lifecycle testing
  - MCP message handling and protocol compliance
  - Resource URI handling (`mindport://` scheme)
  - Tool operations and error scenarios
  - Shorthand domain notation testing

- **`domain_integration_test.go`** - Domain management integration tests
  - Domain creation, hierarchy, and validation
  - Domain isolation and cross-domain access
  - Domain-scoped search operations
  - Shorthand notation parsing and building
  - Domain switching and statistics

- **`workflow_integration_test.go`** - End-to-end workflow tests
  - Complete resource lifecycle (store → index → search → retrieve)
  - Complete prompt lifecycle (store → index → search → execute)
  - Cross-component workflows spanning domains
  - Search optimization and token efficiency
  - Error recovery and batch operations

- **`concurrency_performance_test.go`** - Concurrent access and performance tests
  - Concurrent resource operations (storage, retrieval)
  - Concurrent search operations
  - Mixed read/write workloads
  - Memory usage and resource management
  - Deadlock prevention
  - Performance benchmarks

- **`error_handling_test.go`** - Error handling and edge cases
  - Storage error scenarios
  - Search error handling (invalid queries, malformed regex)
  - Domain operation errors
  - Large data handling
  - Unicode and special character support
  - Boundary value testing
  - Timestamp edge cases

## Running Tests

### Run All Integration Tests
```bash
go test ./tests/... -v
```

### Run Specific Test Files
```bash
# MCP integration tests
go test ./tests/ -run "TestMCP" -v

# Domain tests
go test ./tests/ -run "TestDomain" -v

# Workflow tests
go test ./tests/ -run "TestWorkflow" -v

# Concurrency tests
go test ./tests/ -run "TestConcurrent" -v

# Error handling tests
go test ./tests/ -run "TestError" -v
```

### Run Performance Benchmarks
```bash
# Include performance tests (may take longer)
go test ./tests/ -run "TestPerformance" -v

# Skip performance tests
go test ./tests/ -short -v
```

### Run Tests with Coverage
```bash
go test ./tests/... -coverprofile=coverage.out -v
go tool cover -html=coverage.out -o coverage.html
```

### Run Tests with Race Detection
```bash
go test ./tests/... -race -v
```

## Test Environment

Each test file sets up its own isolated test environment:

- Temporary directories for storage and search indices
- Clean BadgerDB and Bleve instances
- Fresh domain manager configuration
- Automatic cleanup after tests complete

## Test Data

Tests create their own test data including:

- Sample resources with various content types
- Test prompts with template variables
- Multiple domain hierarchies
- Large datasets for performance testing
- Unicode and special character content

## Test Configuration

The tests use the following default configuration:

```go
domainConfig := &domain.DomainConfig{
    DefaultDomain:    "default",
    IsolationMode:    "standard", 
    AllowCrossDomain: true,
}
```

## Known Limitations

Some tests are currently marked as `Skip` because they require:

1. **MCP Protocol Tests**: Direct access to the `handleRequest` method for full protocol testing
2. **Advanced Error Scenarios**: Specific error injection capabilities
3. **Network Tests**: Network-based MCP communication testing

## Performance Expectations

The tests include performance assertions:

- **Storage**: > 50 operations/second
- **Search**: > 20 searches/second  
- **Retrieval**: > 100 retrievals/second
- **Memory**: < 100MB for 1000 1KB resources
- **Concurrent Operations**: < 1% error rate under load

## Contributing

When adding new tests:

1. Follow the existing test structure and naming conventions
2. Create isolated test environments with proper cleanup
3. Include both positive and negative test cases
4. Test error conditions and edge cases
5. Add performance assertions where appropriate
6. Document any test-specific requirements or limitations

## Debugging Tests

For debugging failed tests:

1. Run with verbose output: `-v`
2. Run specific test: `-run "TestName"`
3. Check temporary directories before cleanup
4. Enable race detection: `-race`
5. Use coverage reports to identify untested paths