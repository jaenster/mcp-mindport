#!/bin/bash

# MCP MindPort Test Runner
# This script provides convenient ways to run different test suites

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
VERBOSE=false
COVERAGE=false
RACE=false
SHORT=false
PATTERN=""

# Function to print usage
usage() {
    echo "Usage: $0 [OPTIONS] [TEST_SUITE]"
    echo ""
    echo "Test Suites:"
    echo "  all               Run all tests (default)"
    echo "  unit              Run unit tests only"
    echo "  integration       Run integration tests only"
    echo "  mcp               Run MCP protocol tests"
    echo "  domain            Run domain management tests"
    echo "  workflow          Run workflow tests"
    echo "  concurrency       Run concurrency tests"
    echo "  error             Run error handling tests"
    echo "  performance       Run performance benchmarks"
    echo ""
    echo "Options:"
    echo "  -v, --verbose     Verbose output"
    echo "  -c, --coverage    Generate coverage report"
    echo "  -r, --race        Enable race detection"
    echo "  -s, --short       Skip long-running tests"
    echo "  -p, --pattern     Run tests matching pattern"
    echo "  -h, --help        Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 integration -v          # Run integration tests with verbose output"
    echo "  $0 performance -c          # Run performance tests with coverage"
    echo "  $0 -p \"TestConcurrent\" -r   # Run concurrent tests with race detection"
}

# Function to run tests with given parameters
run_tests() {
    local test_path="$1"
    local test_pattern="$2"
    local description="$3"
    
    echo -e "${BLUE}Running $description...${NC}"
    
    # Build test command
    local cmd="go test"
    
    if [ "$VERBOSE" = true ]; then
        cmd="$cmd -v"
    fi
    
    if [ "$COVERAGE" = true ]; then
        cmd="$cmd -coverprofile=coverage.out"
    fi
    
    if [ "$RACE" = true ]; then
        cmd="$cmd -race"
    fi
    
    if [ "$SHORT" = true ]; then
        cmd="$cmd -short"
    fi
    
    if [ -n "$test_pattern" ]; then
        cmd="$cmd -run $test_pattern"
    fi
    
    cmd="$cmd $test_path"
    
    echo -e "${YELLOW}Command: $cmd${NC}"
    
    # Run the command
    if eval $cmd; then
        echo -e "${GREEN}✓ $description completed successfully${NC}"
        
        # Generate HTML coverage report if coverage was enabled
        if [ "$COVERAGE" = true ] && [ -f "coverage.out" ]; then
            echo -e "${BLUE}Generating coverage report...${NC}"
            go tool cover -html=coverage.out -o coverage.html
            echo -e "${GREEN}✓ Coverage report generated: coverage.html${NC}"
        fi
    else
        echo -e "${RED}✗ $description failed${NC}"
        return 1
    fi
    
    echo ""
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -c|--coverage)
            COVERAGE=true
            shift
            ;;
        -r|--race)
            RACE=true
            shift
            ;;
        -s|--short)
            SHORT=true
            shift
            ;;
        -p|--pattern)
            PATTERN="$2"
            shift 2
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        all|unit|integration|mcp|domain|workflow|concurrency|error|performance)
            TEST_SUITE="$1"
            shift
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            usage
            exit 1
            ;;
    esac
done

# Set default test suite if none specified
if [ -z "$TEST_SUITE" ]; then
    TEST_SUITE="all"
fi

# Override pattern if provided via command line
if [ -n "$PATTERN" ]; then
    TEST_PATTERN="$PATTERN"
fi

echo -e "${BLUE}MCP MindPort Test Runner${NC}"
echo -e "${BLUE}========================${NC}"
echo ""

# Run tests based on suite selection
case $TEST_SUITE in
    all)
        echo -e "${YELLOW}Running all tests...${NC}"
        run_tests "./..." "" "all tests"
        ;;
    unit)
        echo -e "${YELLOW}Running unit tests...${NC}"
        run_tests "./internal/..." "" "unit tests"
        ;;
    integration)
        echo -e "${YELLOW}Running integration tests...${NC}"
        run_tests "./tests/..." "" "integration tests"
        ;;
    mcp)
        echo -e "${YELLOW}Running MCP protocol tests...${NC}"
        run_tests "./tests/" "TestMCP" "MCP protocol tests"
        ;;
    domain)
        echo -e "${YELLOW}Running domain management tests...${NC}"
        run_tests "./tests/" "TestDomain" "domain management tests"
        ;;
    workflow)
        echo -e "${YELLOW}Running workflow tests...${NC}"
        run_tests "./tests/" "TestWorkflow|TestComplete" "workflow tests"
        ;;
    concurrency)
        echo -e "${YELLOW}Running concurrency tests...${NC}"
        run_tests "./tests/" "TestConcurrent|TestMixed|TestMemory|TestDeadlock" "concurrency tests"
        ;;
    error)
        echo -e "${YELLOW}Running error handling tests...${NC}"
        run_tests "./tests/" "TestError|TestEdge" "error handling tests"
        ;;
    performance)
        echo -e "${YELLOW}Running performance benchmarks...${NC}"
        # Remove short flag for performance tests
        if [ "$SHORT" = true ]; then
            echo -e "${YELLOW}Note: Ignoring --short flag for performance tests${NC}"
            SHORT=false
        fi
        run_tests "./tests/" "TestPerformance|TestBatch|TestLarge" "performance benchmarks"
        ;;
    *)
        echo -e "${RED}Unknown test suite: $TEST_SUITE${NC}"
        usage
        exit 1
        ;;
esac

echo -e "${GREEN}Test run completed!${NC}"

# Show coverage summary if coverage was enabled
if [ "$COVERAGE" = true ] && [ -f "coverage.out" ]; then
    echo ""
    echo -e "${BLUE}Coverage Summary:${NC}"
    go tool cover -func=coverage.out | tail -1
fi