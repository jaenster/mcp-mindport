package search

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2"
	blevequery "github.com/blevesearch/bleve/v2/search/query"
)

// AdvancedSearchQuery represents a sophisticated search query with multiple options
type AdvancedSearchQuery struct {
	// Basic search
	Query string `json:"query"`
	
	// Search modes
	Mode        string `json:"mode,omitempty"`        // "exact", "fuzzy", "regex", "wildcard", "semantic"
	CaseSensitive bool `json:"case_sensitive,omitempty"`
	WholeWords    bool `json:"whole_words,omitempty"`
	
	// Content filters
	Type         string   `json:"type,omitempty"`         // "resource", "prompt"
	Tags         []string `json:"tags,omitempty"`
	ContentType  string   `json:"content_type,omitempty"` // "code", "documentation", "data", etc.
	Domains      []string `json:"domains,omitempty"`      // Domain filtering
	
	// Date filters
	CreatedAfter  *time.Time `json:"created_after,omitempty"`
	CreatedBefore *time.Time `json:"created_before,omitempty"`
	UpdatedAfter  *time.Time `json:"updated_after,omitempty"`
	UpdatedBefore *time.Time `json:"updated_before,omitempty"`
	
	// Advanced filters
	MinScore    float64 `json:"min_score,omitempty"`
	HasMetadata bool    `json:"has_metadata,omitempty"`
	
	// Field-specific search
	Fields []string `json:"fields,omitempty"` // "title", "content", "tags", "metadata"
	
	// CLI-style options
	Include []string `json:"include,omitempty"` // Include patterns (like grep -e)
	Exclude []string `json:"exclude,omitempty"` // Exclude patterns (like grep -v)
	Context int      `json:"context,omitempty"` // Lines of context (like grep -C)
	
	// Result options
	Limit      int    `json:"limit,omitempty"`
	Offset     int    `json:"offset,omitempty"`
	SortBy     string `json:"sort_by,omitempty"`     // "relevance", "date", "title", "type"
	SortOrder  string `json:"sort_order,omitempty"`  // "asc", "desc"
	Highlight  bool   `json:"highlight,omitempty"`
	SnippetLen int    `json:"snippet_len,omitempty"`
}

// SearchResult represents an enhanced search result
type AdvancedSearchResult struct {
	ID          string                 `json:"id"`
	Score       float64                `json:"score"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Metadata    map[string]interface{} `json:"metadata"`
	Tags        []string               `json:"tags"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	
	// Enhanced result info
	Snippet     string            `json:"snippet"`
	Highlights  []string          `json:"highlights,omitempty"`
	ContextLines []string         `json:"context_lines,omitempty"`
	FieldMatches map[string]int   `json:"field_matches,omitempty"`
	
	// CLI-style info
	LineNumbers []int `json:"line_numbers,omitempty"`
	MatchCount  int   `json:"match_count"`
}

// SearchStats provides search analytics
type SearchStats struct {
	TotalResults    int                    `json:"total_results"`
	SearchTime      time.Duration          `json:"search_time"`
	TypeBreakdown   map[string]int         `json:"type_breakdown"`
	TagBreakdown    map[string]int         `json:"tag_breakdown"`
	ScoreDistribution map[string]int       `json:"score_distribution"`
	FieldMatches    map[string]int         `json:"field_matches"`
}

// AdvancedSearch performs sophisticated search with multiple modes
func (s *BleveSearch) AdvancedSearch(ctx context.Context, query *AdvancedSearchQuery) ([]*AdvancedSearchResult, *SearchStats, error) {
	startTime := time.Now()
	
	if query.Limit <= 0 {
		query.Limit = 20
	}
	if query.SnippetLen <= 0 {
		query.SnippetLen = 200
	}

	// Build the search query based on mode
	bleveQuery, err := s.buildAdvancedQuery(query)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build query: %w", err)
	}

	// Create search request
	searchRequest := bleve.NewSearchRequest(bleveQuery)
	searchRequest.Size = query.Limit
	searchRequest.From = query.Offset
	searchRequest.Fields = []string{"*"}
	
	// Add highlighting if requested
	if query.Highlight {
		highlight := bleve.NewHighlight()
		highlight.AddField("title")
		highlight.AddField("content")
		searchRequest.Highlight = highlight
	}

	// Execute search
	searchResult, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, nil, fmt.Errorf("search failed: %w", err)
	}

	// Convert and enhance results
	results, stats := s.processAdvancedResults(searchResult, query, time.Since(startTime))
	
	// Apply post-processing filters
	results = s.applyPostFilters(results, query)
	
	// Sort results if requested
	if query.SortBy != "" {
		s.sortResults(results, query.SortBy, query.SortOrder)
	}

	return results, stats, nil
}

// buildAdvancedQuery constructs a Bleve query based on the search mode
func (s *BleveSearch) buildAdvancedQuery(query *AdvancedSearchQuery) (blevequery.Query, error) {
	var baseQuery blevequery.Query
	
	// If no query text, start with match all and only use filters
	if strings.TrimSpace(query.Query) == "" {
		baseQuery = bleve.NewMatchAllQuery()
	} else {
		switch query.Mode {
		case "exact":
			baseQuery = s.buildExactQuery(query)
		case "fuzzy":
			baseQuery = s.buildFuzzyQuery(query)
		case "regex":
			baseQuery = s.buildRegexQuery(query)
		case "wildcard":
			baseQuery = s.buildWildcardQuery(query)
		case "semantic":
			baseQuery = s.buildSemanticQuery(query)
		default:
			baseQuery = s.buildSmartQuery(query) // Auto-detect best mode
		}
	}

	// Apply filters
	filters := s.buildFilters(query)
	if len(filters) > 0 {
		if strings.TrimSpace(query.Query) == "" {
			// If no base query, use only filters
			if len(filters) == 1 {
				baseQuery = filters[0]
			} else {
				baseQuery = bleve.NewConjunctionQuery(filters...)
			}
		} else {
			// Combine base query with filters
			filters = append(filters, baseQuery)
			baseQuery = bleve.NewConjunctionQuery(filters...)
		}
	}

	return baseQuery, nil
}

// buildSmartQuery automatically chooses the best search mode
func (s *BleveSearch) buildSmartQuery(query *AdvancedSearchQuery) blevequery.Query {
	q := query.Query
	
	// Detect patterns and choose appropriate mode
	if strings.Contains(q, "*") || strings.Contains(q, "?") {
		return s.buildWildcardQuery(query)
	}
	if strings.HasPrefix(q, "/") && strings.HasSuffix(q, "/") {
		return s.buildRegexQuery(query)
	}
	if strings.Contains(q, "~") {
		return s.buildFuzzyQuery(query)
	}
	if len(strings.Fields(q)) > 3 {
		return s.buildSemanticQuery(query)
	}
	
	// Default to multi-field search with boosting
	return s.buildMultiFieldQuery(query)
}

// buildMultiFieldQuery creates a boosted multi-field search
func (s *BleveSearch) buildMultiFieldQuery(query *AdvancedSearchQuery) blevequery.Query {
	q := query.Query
	if !query.CaseSensitive {
		q = strings.ToLower(q)
	}

	queries := []blevequery.Query{}
	
	// Title search (highest boost)
	titleQuery := bleve.NewMatchQuery(q)
	titleQuery.SetField("title")
	titleQuery.SetBoost(5.0)
	queries = append(queries, titleQuery)
	
	// Content search (medium boost)
	contentQuery := bleve.NewMatchQuery(q)
	contentQuery.SetField("content")
	contentQuery.SetBoost(2.0)
	queries = append(queries, contentQuery)
	
	// Tags search (high boost)
	tagsQuery := bleve.NewMatchQuery(q)
	tagsQuery.SetField("tags")
	tagsQuery.SetBoost(3.0)
	queries = append(queries, tagsQuery)
	
	// Search terms (medium boost)
	searchTermsQuery := bleve.NewMatchQuery(q)
	searchTermsQuery.SetField("search_terms")
	searchTermsQuery.SetBoost(2.5)
	queries = append(queries, searchTermsQuery)

	return bleve.NewDisjunctionQuery(queries...)
}

// buildExactQuery creates an exact phrase search
func (s *BleveSearch) buildExactQuery(query *AdvancedSearchQuery) blevequery.Query {
	q := query.Query
	if !query.CaseSensitive {
		q = strings.ToLower(q)
	}
	
	return bleve.NewMatchPhraseQuery(q)
}

// buildFuzzyQuery creates a fuzzy search for typo tolerance
func (s *BleveSearch) buildFuzzyQuery(query *AdvancedSearchQuery) blevequery.Query {
	q := strings.TrimSuffix(query.Query, "~")
	fuzzyQuery := bleve.NewFuzzyQuery(q)
	return fuzzyQuery
}

// buildRegexQuery creates a regex-based search
func (s *BleveSearch) buildRegexQuery(query *AdvancedSearchQuery) blevequery.Query {
	q := query.Query
	if strings.HasPrefix(q, "/") && strings.HasSuffix(q, "/") {
		q = q[1 : len(q)-1] // Remove regex delimiters
	}
	
	regexQuery := bleve.NewRegexpQuery(q)
	return regexQuery
}

// buildWildcardQuery creates a wildcard search
func (s *BleveSearch) buildWildcardQuery(query *AdvancedSearchQuery) blevequery.Query {
	wildcardQuery := bleve.NewWildcardQuery(query.Query)
	return wildcardQuery
}

// buildSemanticQuery creates a semantic/conceptual search
func (s *BleveSearch) buildSemanticQuery(query *AdvancedSearchQuery) blevequery.Query {
	// For semantic search, we use a combination of techniques
	words := strings.Fields(query.Query)
	queries := []blevequery.Query{}
	
	for _, word := range words {
		// Add the word itself
		exactQuery := bleve.NewMatchQuery(word)
		exactQuery.SetBoost(2.0)
		queries = append(queries, exactQuery)
		
		// Add fuzzy variant for related terms
		fuzzyQuery := bleve.NewFuzzyQuery(word)
		fuzzyQuery.SetBoost(0.5)
		queries = append(queries, fuzzyQuery)
	}
	
	// Require at least half the terms to match
	disjunctionQuery := bleve.NewDisjunctionQuery(queries...)
	minMatches := len(words) / 2
	if minMatches > 0 {
		disjunctionQuery.SetMin(float64(minMatches))
	}
	
	return disjunctionQuery
}

// buildFilters creates filters for type, tags, dates, etc.
func (s *BleveSearch) buildFilters(query *AdvancedSearchQuery) []blevequery.Query {
	var filters []blevequery.Query
	
	// Type filter
	if query.Type != "" {
		typeQuery := bleve.NewTermQuery(query.Type)
		typeQuery.SetField("type")
		filters = append(filters, typeQuery)
	}
	
	// Tags filter
	if len(query.Tags) > 0 {
		tagQueries := []blevequery.Query{}
		for _, tag := range query.Tags {
			tagQuery := bleve.NewMatchQuery(tag)
			tagQuery.SetField("tags")
			tagQueries = append(tagQueries, tagQuery)
		}
		filters = append(filters, bleve.NewDisjunctionQuery(tagQueries...))
	}
	
	// Content type filter
	if query.ContentType != "" {
		contentTypeQuery := bleve.NewTermQuery(query.ContentType)
		contentTypeQuery.SetField("content_type")
		filters = append(filters, contentTypeQuery)
	}
	
	// Date range filters would be added here
	// (requires date field mapping in the index)
	
	return filters
}

// processAdvancedResults converts Bleve results to enhanced results
func (s *BleveSearch) processAdvancedResults(searchResult *bleve.SearchResult, query *AdvancedSearchQuery, searchTime time.Duration) ([]*AdvancedSearchResult, *SearchStats) {
	results := []*AdvancedSearchResult{}
	stats := &SearchStats{
		TotalResults:      int(searchResult.Total),
		SearchTime:        searchTime,
		TypeBreakdown:     make(map[string]int),
		TagBreakdown:      make(map[string]int),
		ScoreDistribution: make(map[string]int),
		FieldMatches:      make(map[string]int),
	}
	
	for _, hit := range searchResult.Hits {
		result := &AdvancedSearchResult{
			ID:           hit.ID,
			Score:        hit.Score,
			FieldMatches: make(map[string]int),
		}
		
		// Extract basic fields
		if title, ok := hit.Fields["title"].(string); ok {
			result.Title = title
		}
		if content, ok := hit.Fields["content"].(string); ok {
			result.Content = content
			result.Snippet = s.createSnippet(content, query.Query, query.SnippetLen)
		}
		if docType, ok := hit.Fields["type"].(string); ok {
			result.Type = docType
			stats.TypeBreakdown[docType]++
		}
		if tags, ok := hit.Fields["tags"].(string); ok {
			result.Tags = strings.Fields(tags)
			for _, tag := range result.Tags {
				stats.TagBreakdown[tag]++
			}
		}
		
		// Add highlights if available
		if hit.Fragments != nil {
			for field, fragments := range hit.Fragments {
				result.Highlights = append(result.Highlights, fragments...)
				result.FieldMatches[field] = len(fragments)
				stats.FieldMatches[field]++
			}
		}
		
		// Calculate match count and line numbers for CLI-style output
		result.MatchCount = s.countMatches(result.Content, query.Query)
		result.LineNumbers = s.findMatchingLines(result.Content, query.Query)
		
		// Add context lines if requested
		if query.Context > 0 {
			result.ContextLines = s.getContextLines(result.Content, result.LineNumbers, query.Context)
		}
		
		// Score distribution for analytics
		scoreRange := fmt.Sprintf("%.1f-%.1f", 
			float64(int(result.Score*10))/10, 
			float64(int(result.Score*10)+1)/10)
		stats.ScoreDistribution[scoreRange]++
		
		results = append(results, result)
	}
	
	return results, stats
}

// createSnippet creates a context-aware snippet around matches
func (s *BleveSearch) createSnippet(content, query string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	
	// Find the first occurrence of any query term
	queryTerms := strings.Fields(strings.ToLower(query))
	contentLower := strings.ToLower(content)
	
	bestPos := -1
	for _, term := range queryTerms {
		if pos := strings.Index(contentLower, term); pos >= 0 {
			if bestPos == -1 || pos < bestPos {
				bestPos = pos
			}
		}
	}
	
	if bestPos == -1 {
		return content[:maxLen] + "..."
	}
	
	// Center the snippet around the match
	start := bestPos - maxLen/4
	if start < 0 {
		start = 0
	}
	
	end := start + maxLen
	if end > len(content) {
		end = len(content)
		start = end - maxLen
		if start < 0 {
			start = 0
		}
	}
	
	snippet := content[start:end]
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(content) {
		snippet = snippet + "..."
	}
	
	return snippet
}

// countMatches counts occurrences of query terms in content
func (s *BleveSearch) countMatches(content, query string) int {
	queryTerms := strings.Fields(strings.ToLower(query))
	contentLower := strings.ToLower(content)
	
	count := 0
	for _, term := range queryTerms {
		count += strings.Count(contentLower, term)
	}
	
	return count
}

// findMatchingLines finds line numbers that contain matches
func (s *BleveSearch) findMatchingLines(content, query string) []int {
	lines := strings.Split(content, "\n")
	queryTerms := strings.Fields(strings.ToLower(query))
	matchingLines := []int{}
	
	for i, line := range lines {
		lineLower := strings.ToLower(line)
		for _, term := range queryTerms {
			if strings.Contains(lineLower, term) {
				matchingLines = append(matchingLines, i+1) // 1-based line numbers
				break
			}
		}
	}
	
	return matchingLines
}

// getContextLines extracts context lines around matches
func (s *BleveSearch) getContextLines(content string, matchingLines []int, context int) []string {
	if len(matchingLines) == 0 {
		return []string{}
	}
	
	lines := strings.Split(content, "\n")
	contextLines := []string{}
	
	for _, lineNum := range matchingLines {
		// Add context before
		for i := lineNum - context - 1; i < lineNum-1; i++ {
			if i >= 0 && i < len(lines) {
				contextLines = append(contextLines, fmt.Sprintf("%d-  %s", i+1, lines[i]))
			}
		}
		
		// Add the matching line
		if lineNum-1 >= 0 && lineNum-1 < len(lines) {
			contextLines = append(contextLines, fmt.Sprintf("%d:  %s", lineNum, lines[lineNum-1]))
		}
		
		// Add context after
		for i := lineNum; i < lineNum+context; i++ {
			if i >= 0 && i < len(lines) {
				contextLines = append(contextLines, fmt.Sprintf("%d-  %s", i+1, lines[i]))
			}
		}
	}
	
	return contextLines
}

// applyPostFilters applies filters that can't be done at the Bleve level
func (s *BleveSearch) applyPostFilters(results []*AdvancedSearchResult, query *AdvancedSearchQuery) []*AdvancedSearchResult {
	filtered := []*AdvancedSearchResult{}
	
	for _, result := range results {
		// Score filter
		if query.MinScore > 0 && result.Score < query.MinScore {
			continue
		}
		
		// Include/Exclude patterns (CLI-style)
		if len(query.Include) > 0 {
			matches := false
			for _, pattern := range query.Include {
				if s.matchesPattern(result.Content, pattern) {
					matches = true
					break
				}
			}
			if !matches {
				continue
			}
		}
		
		if len(query.Exclude) > 0 {
			exclude := false
			for _, pattern := range query.Exclude {
				if s.matchesPattern(result.Content, pattern) {
					exclude = true
					break
				}
			}
			if exclude {
				continue
			}
		}
		
		filtered = append(filtered, result)
	}
	
	return filtered
}

// matchesPattern checks if content matches a pattern (supports regex)
func (s *BleveSearch) matchesPattern(content, pattern string) bool {
	// Try regex first
	if regexp, err := regexp.Compile(pattern); err == nil {
		return regexp.MatchString(content)
	}
	
	// Fallback to simple substring match
	return strings.Contains(strings.ToLower(content), strings.ToLower(pattern))
}

// sortResults sorts results by the specified criteria
func (s *BleveSearch) sortResults(results []*AdvancedSearchResult, sortBy, sortOrder string) {
	sort.Slice(results, func(i, j int) bool {
		var less bool
		
		switch sortBy {
		case "relevance", "score":
			less = results[i].Score > results[j].Score // Higher score first
		case "date", "updated":
			less = results[i].UpdatedAt.After(results[j].UpdatedAt)
		case "created":
			less = results[i].CreatedAt.After(results[j].CreatedAt)
		case "title":
			less = results[i].Title < results[j].Title
		case "type":
			less = results[i].Type < results[j].Type
		default:
			less = results[i].Score > results[j].Score
		}
		
		if sortOrder == "asc" {
			return !less
		}
		return less
	})
}