package search

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"mcp-mindport/internal/storage"
)

// CLISearchTools provides familiar CLI-style search commands
type CLISearchTools struct {
	storage      *storage.BadgerStore
	searchEngine *BleveSearch
}

// GrepOptions represents options for grep-like search
type GrepOptions struct {
	Pattern       string   `json:"pattern"`
	IgnoreCase    bool     `json:"ignore_case,omitempty"`    // -i
	InvertMatch   bool     `json:"invert_match,omitempty"`   // -v
	LineNumbers   bool     `json:"line_numbers,omitempty"`   // -n
	Count         bool     `json:"count,omitempty"`          // -c
	Context       int      `json:"context,omitempty"`        // -C
	ContextAfter  int      `json:"context_after,omitempty"`  // -A
	ContextBefore int      `json:"context_before,omitempty"` // -B
	Recursive     bool     `json:"recursive,omitempty"`      // -r
	Include       []string `json:"include,omitempty"`        // --include
	Exclude       []string `json:"exclude,omitempty"`        // --exclude
	MaxMatches    int      `json:"max_matches,omitempty"`    // -m
	WholeWords    bool     `json:"whole_words,omitempty"`    // -w
	Extended      bool     `json:"extended,omitempty"`       // -E (regex)
	Fixed         bool     `json:"fixed,omitempty"`          // -F (fixed strings)
	OnlyMatching  bool     `json:"only_matching,omitempty"`  // -o
}

// FindOptions represents options for find-like search
type FindOptions struct {
	Name         string    `json:"name,omitempty"`          // -name
	Type         string    `json:"type,omitempty"`          // -type (f=file, d=directory)
	Size         string    `json:"size,omitempty"`          // -size (+100k, -1M)
	Modified     string    `json:"modified,omitempty"`      // -mtime, -newer
	Created      string    `json:"created,omitempty"`       // -ctime
	Tags         []string  `json:"tags,omitempty"`          // custom: search by tags
	ContentType  string    `json:"content_type,omitempty"`  // custom: filter by content type
	MaxDepth     int       `json:"max_depth,omitempty"`     // -maxdepth
	MinDepth     int       `json:"min_depth,omitempty"`     // -mindepth
	Execute      string    `json:"execute,omitempty"`       // -exec
	Print        bool      `json:"print,omitempty"`         // -print
	Limit        int       `json:"limit,omitempty"`
}

// RipgrepOptions represents options for ripgrep-style search (more advanced than grep)
type RipgrepOptions struct {
	Pattern       string   `json:"pattern"`
	IgnoreCase    bool     `json:"ignore_case,omitempty"`     // -i
	SmartCase     bool     `json:"smart_case,omitempty"`      // -S
	Multiline     bool     `json:"multiline,omitempty"`       // -U
	DotAll        bool     `json:"dot_all,omitempty"`         // -s
	CaseSensitive bool     `json:"case_sensitive,omitempty"`  // -s
	WordRegexp    bool     `json:"word_regexp,omitempty"`     // -w
	LineRegexp    bool     `json:"line_regexp,omitempty"`     // -x
	Fixed         bool     `json:"fixed,omitempty"`           // -F
	InvertMatch   bool     `json:"invert_match,omitempty"`    // -v
	Count         bool     `json:"count,omitempty"`           // -c
	CountMatches  bool     `json:"count_matches,omitempty"`   // --count-matches
	FilesWithMatches bool  `json:"files_with_matches,omitempty"` // -l
	FilesWithoutMatch bool `json:"files_without_match,omitempty"` // -L
	WithFilename  bool     `json:"with_filename,omitempty"`   // -H
	NoFilename    bool     `json:"no_filename,omitempty"`     // -h
	LineNumber    bool     `json:"line_number,omitempty"`     // -n
	NoLineNumber  bool     `json:"no_line_number,omitempty"`  // -N
	OnlyMatching  bool     `json:"only_matching,omitempty"`   // -o
	Replace       string   `json:"replace,omitempty"`         // -r
	Context       int      `json:"context,omitempty"`         // -C
	ContextAfter  int      `json:"context_after,omitempty"`   // -A
	ContextBefore int      `json:"context_before,omitempty"`  // -B
	MaxCount      int      `json:"max_count,omitempty"`       // -m
	MaxFileSize   string   `json:"max_file_size,omitempty"`   // --max-filesize
	Include       []string `json:"include,omitempty"`         // -g
	Exclude       []string `json:"exclude,omitempty"`         // -g !pattern
	Type          []string `json:"type,omitempty"`            // -t
	TypeNot       []string `json:"type_not,omitempty"`        // -T
}

// GrepResult represents a grep-style search result
type GrepResult struct {
	ResourceID   string   `json:"resource_id"`
	ResourceType string   `json:"resource_type"`
	Title        string   `json:"title"`
	LineNumber   int      `json:"line_number,omitempty"`
	MatchedLine  string   `json:"matched_line"`
	Context      []string `json:"context,omitempty"`
	MatchCount   int      `json:"match_count"`
}

// FindResult represents a find-style search result
type FindResult struct {
	ResourceID   string                 `json:"resource_id"`
	ResourceType string                 `json:"resource_type"`
	Title        string                 `json:"title"`
	Size         int64                  `json:"size"`
	Created      time.Time              `json:"created"`
	Modified     time.Time              `json:"modified"`
	Tags         []string               `json:"tags"`
	Metadata     map[string]interface{} `json:"metadata"`
	Path         string                 `json:"path"` // Virtual path for resources
}

func NewCLISearchTools(storage *storage.BadgerStore, searchEngine *BleveSearch) *CLISearchTools {
	return &CLISearchTools{
		storage:      storage,
		searchEngine: searchEngine,
	}
}

// Grep performs grep-like search across resources
func (c *CLISearchTools) Grep(ctx context.Context, opts *GrepOptions) ([]*GrepResult, error) {
	if opts.MaxMatches == 0 {
		opts.MaxMatches = 1000
	}

	// Get all resources
	resources, err := c.storage.ListResources(ctx, 10000, 0) // Get a large batch
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}

	var results []*GrepResult
	totalMatches := 0

	for _, resource := range resources {
		if totalMatches >= opts.MaxMatches {
			break
		}

		// Apply include/exclude filters
		if !c.shouldIncludeResource(resource, opts.Include, opts.Exclude) {
			continue
		}

		// Search in the resource content
		matches := c.grepInContent(resource, opts)
		results = append(results, matches...)
		totalMatches += len(matches)
	}

	// Apply count-only option
	if opts.Count {
		// Return summary result
		countResult := &GrepResult{
			ResourceID:  "summary",
			Title:       "Total matches",
			MatchCount:  totalMatches,
			MatchedLine: fmt.Sprintf("%d total matches across %d resources", totalMatches, len(results)),
		}
		return []*GrepResult{countResult}, nil
	}

	return results, nil
}

// Find performs find-like search for resources
func (c *CLISearchTools) Find(ctx context.Context, opts *FindOptions) ([]*FindResult, error) {
	if opts.Limit == 0 {
		opts.Limit = 1000
	}

	// Get all resources and prompts
	resources, err := c.storage.ListResources(ctx, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}

	prompts, err := c.storage.ListPrompts(ctx, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list prompts: %w", err)
	}

	var results []*FindResult

	// Process resources
	for _, resource := range resources {
		if len(results) >= opts.Limit {
			break
		}

		if c.matchesFindCriteria(resource, nil, opts) {
			results = append(results, &FindResult{
				ResourceID:   resource.ID,
				ResourceType: "resource",
				Title:        resource.Title,
				Size:         int64(len(resource.Content)),
				Created:      resource.CreatedAt,
				Modified:     resource.UpdatedAt,
				Tags:         resource.Tags,
				Metadata:     resource.Metadata,
				Path:         fmt.Sprintf("/resources/%s", resource.ID),
			})
		}
	}

	// Process prompts
	for _, prompt := range prompts {
		if len(results) >= opts.Limit {
			break
		}

		if c.matchesFindCriteria(nil, prompt, opts) {
			results = append(results, &FindResult{
				ResourceID:   prompt.ID,
				ResourceType: "prompt",
				Title:        prompt.Name,
				Size:         int64(len(prompt.Template)),
				Created:      prompt.CreatedAt,
				Modified:     prompt.UpdatedAt,
				Tags:         prompt.Tags,
				Path:         fmt.Sprintf("/prompts/%s", prompt.ID),
			})
		}
	}

	return results, nil
}

// Ripgrep performs ripgrep-style advanced search
func (c *CLISearchTools) Ripgrep(ctx context.Context, opts *RipgrepOptions) ([]*GrepResult, error) {
	// Convert ripgrep options to advanced search query
	query := &AdvancedSearchQuery{
		Query:         opts.Pattern,
		CaseSensitive: opts.CaseSensitive || (!opts.IgnoreCase && !opts.SmartCase),
		WholeWords:    opts.WordRegexp,
		Limit:         opts.MaxCount,
		Context:       opts.Context,
		Highlight:     true,
		Fields:        []string{"title", "content"},
	}

	// Handle smart case
	if opts.SmartCase && c.hasUpperCase(opts.Pattern) {
		query.CaseSensitive = true
	}

	// Set search mode based on options
	if opts.Fixed {
		query.Mode = "exact"
	} else if opts.Multiline || opts.DotAll {
		query.Mode = "regex"
	} else {
		query.Mode = "smart"
	}

	// Apply type filters
	if len(opts.Type) > 0 {
		query.Type = opts.Type[0] // Use first type for simplicity
	}

	// Perform advanced search
	searchResults, _, err := c.searchEngine.AdvancedSearch(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ripgrep search failed: %w", err)
	}

	// Convert to grep-style results
	var results []*GrepResult
	for _, result := range searchResults {
		grepResult := &GrepResult{
			ResourceID:   result.ID,
			ResourceType: result.Type,
			Title:        result.Title,
			MatchedLine:  result.Snippet,
			MatchCount:   result.MatchCount,
			Context:      result.ContextLines,
		}

		if len(result.LineNumbers) > 0 {
			grepResult.LineNumber = result.LineNumbers[0]
		}

		results = append(results, grepResult)
	}

	// Apply special formatting options
	if opts.Count {
		totalMatches := 0
		for _, result := range results {
			totalMatches += result.MatchCount
		}
		return []*GrepResult{{
			ResourceID:  "summary",
			Title:       "Total matches",
			MatchCount:  totalMatches,
			MatchedLine: fmt.Sprintf("%d matches", totalMatches),
		}}, nil
	}

	if opts.FilesWithMatches {
		// Return only unique resources that have matches
		uniqueResults := make(map[string]*GrepResult)
		for _, result := range results {
			if _, exists := uniqueResults[result.ResourceID]; !exists {
				uniqueResults[result.ResourceID] = &GrepResult{
					ResourceID:   result.ResourceID,
					ResourceType: result.ResourceType,
					Title:        result.Title,
					MatchedLine:  fmt.Sprintf("Resource: %s", result.Title),
					MatchCount:   1,
				}
			}
		}
		
		var fileResults []*GrepResult
		for _, result := range uniqueResults {
			fileResults = append(fileResults, result)
		}
		return fileResults, nil
	}

	return results, nil
}

// grepInContent searches for pattern within resource content
func (c *CLISearchTools) grepInContent(resource *storage.Resource, opts *GrepOptions) []*GrepResult {
	content := resource.Content
	lines := strings.Split(content, "\n")
	var results []*GrepResult

	pattern := opts.Pattern
	if opts.IgnoreCase {
		pattern = strings.ToLower(pattern)
	}

	// Compile regex if extended mode
	var regex *regexp.Regexp
	if opts.Extended && !opts.Fixed {
		var err error
		flags := ""
		if opts.IgnoreCase {
			flags += "(?i)"
		}
		regex, err = regexp.Compile(flags + pattern)
		if err != nil {
			// Fallback to literal search
			regex = nil
		}
	}

	for i, line := range lines {
		var matches bool
		searchLine := line
		if opts.IgnoreCase && regex == nil {
			searchLine = strings.ToLower(line)
		}

		// Determine if line matches
		if regex != nil {
			matches = regex.MatchString(line)
		} else if opts.WholeWords {
			matches = c.matchWholeWords(searchLine, pattern)
		} else if opts.Fixed {
			matches = strings.Contains(searchLine, pattern)
		} else {
			matches = strings.Contains(searchLine, pattern)
		}

		// Apply invert match
		if opts.InvertMatch {
			matches = !matches
		}

		if matches {
			result := &GrepResult{
				ResourceID:   resource.ID,
				ResourceType: resource.Type,
				Title:        resource.Title,
				MatchedLine:  line,
				MatchCount:   1,
			}

			if opts.LineNumbers {
				result.LineNumber = i + 1
			}

			// Add context if requested
			if opts.Context > 0 || opts.ContextBefore > 0 || opts.ContextAfter > 0 {
				result.Context = c.getContextLines(lines, i, opts)
			}

			// Only show matching part if requested
			if opts.OnlyMatching {
				if regex != nil {
					if match := regex.FindString(line); match != "" {
						result.MatchedLine = match
					}
				} else {
					// Find and extract the matching substring
					if pos := strings.Index(searchLine, pattern); pos >= 0 {
						end := pos + len(pattern)
						if end <= len(line) {
							result.MatchedLine = line[pos:end]
						}
					}
				}
			}

			results = append(results, result)
		}
	}

	return results
}

// shouldIncludeResource checks if resource should be included based on filters
func (c *CLISearchTools) shouldIncludeResource(resource *storage.Resource, include, exclude []string) bool {
	// Check include patterns
	if len(include) > 0 {
		included := false
		for _, pattern := range include {
			if c.matchesResourcePattern(resource, pattern) {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}

	// Check exclude patterns
	for _, pattern := range exclude {
		if c.matchesResourcePattern(resource, pattern) {
			return false
		}
	}

	return true
}

// matchesResourcePattern checks if resource matches a pattern
func (c *CLISearchTools) matchesResourcePattern(resource *storage.Resource, pattern string) bool {
	// Check title
	if strings.Contains(strings.ToLower(resource.Title), strings.ToLower(pattern)) {
		return true
	}
	
	// Check type
	if strings.Contains(strings.ToLower(resource.Type), strings.ToLower(pattern)) {
		return true
	}
	
	// Check tags
	for _, tag := range resource.Tags {
		if strings.Contains(strings.ToLower(tag), strings.ToLower(pattern)) {
			return true
		}
	}
	
	return false
}

// matchesFindCriteria checks if resource/prompt matches find criteria
func (c *CLISearchTools) matchesFindCriteria(resource *storage.Resource, prompt *storage.Prompt, opts *FindOptions) bool {
	// Name pattern matching
	if opts.Name != "" {
		name := ""
		if resource != nil {
			name = resource.Title
		} else if prompt != nil {
			name = prompt.Name
		}
		
		if matched, _ := regexp.MatchString(opts.Name, name); !matched {
			return false
		}
	}

	// Type filtering
	if opts.Type != "" {
		if resource != nil {
			if opts.Type != "f" && opts.Type != "file" {
				return false
			}
		} else if prompt != nil {
			if opts.Type != "d" && opts.Type != "directory" && opts.Type != "prompt" {
				return false
			}
		}
	}

	// Content type filtering
	if opts.ContentType != "" {
		if resource != nil && resource.Type != opts.ContentType {
			return false
		}
	}

	// Tags filtering
	if len(opts.Tags) > 0 {
		resourceTags := []string{}
		if resource != nil {
			resourceTags = resource.Tags
		} else if prompt != nil {
			resourceTags = prompt.Tags
		}
		
		hasMatchingTag := false
		for _, requiredTag := range opts.Tags {
			for _, resourceTag := range resourceTags {
				if strings.EqualFold(requiredTag, resourceTag) {
					hasMatchingTag = true
					break
				}
			}
			if hasMatchingTag {
				break
			}
		}
		if !hasMatchingTag {
			return false
		}
	}

	// Size filtering (simplified)
	if opts.Size != "" {
		size := int64(0)
		if resource != nil {
			size = int64(len(resource.Content))
		} else if prompt != nil {
			size = int64(len(prompt.Template))
		}
		
		if !c.matchesSizeFilter(size, opts.Size) {
			return false
		}
	}

	return true
}

// matchesSizeFilter checks if size matches the filter (+100k, -1M, etc.)
func (c *CLISearchTools) matchesSizeFilter(size int64, filter string) bool {
	if len(filter) < 2 {
		return true
	}
	
	operator := filter[0]
	sizeStr := filter[1:]
	
	// Parse size with suffix
	multiplier := int64(1)
	if strings.HasSuffix(sizeStr, "k") || strings.HasSuffix(sizeStr, "K") {
		multiplier = 1024
		sizeStr = sizeStr[:len(sizeStr)-1]
	} else if strings.HasSuffix(sizeStr, "m") || strings.HasSuffix(sizeStr, "M") {
		multiplier = 1024 * 1024
		sizeStr = sizeStr[:len(sizeStr)-1]
	} else if strings.HasSuffix(sizeStr, "g") || strings.HasSuffix(sizeStr, "G") {
		multiplier = 1024 * 1024 * 1024
		sizeStr = sizeStr[:len(sizeStr)-1]
	}
	
	targetSize, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return true
	}
	targetSize *= multiplier
	
	switch operator {
	case '+':
		return size > targetSize
	case '-':
		return size < targetSize
	default:
		return size == targetSize
	}
}

// matchWholeWords checks for whole word matches
func (c *CLISearchTools) matchWholeWords(text, pattern string) bool {
	words := strings.Fields(text)
	for _, word := range words {
		if word == pattern {
			return true
		}
	}
	return false
}

// getContextLines extracts context lines around a match
func (c *CLISearchTools) getContextLines(lines []string, matchIndex int, opts *GrepOptions) []string {
	var context []string
	
	before := opts.ContextBefore
	after := opts.ContextAfter
	if opts.Context > 0 {
		before = opts.Context
		after = opts.Context
	}
	
	start := matchIndex - before
	if start < 0 {
		start = 0
	}
	
	end := matchIndex + after + 1
	if end > len(lines) {
		end = len(lines)
	}
	
	for i := start; i < end; i++ {
		prefix := "  "
		if i == matchIndex {
			prefix = "> "
		}
		context = append(context, fmt.Sprintf("%s%d: %s", prefix, i+1, lines[i]))
	}
	
	return context
}

// hasUpperCase checks if string contains uppercase letters
func (c *CLISearchTools) hasUpperCase(s string) bool {
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			return true
		}
	}
	return false
}