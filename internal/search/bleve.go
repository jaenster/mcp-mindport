package search

import (
	"context"
	"fmt"
	"strings"

	"mcp-mindport/internal/storage"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	blevequery "github.com/blevesearch/bleve/v2/search/query"
)

type BleveSearch struct {
	index bleve.Index
}

type SearchResult struct {
	ID       string                 `json:"id"`
	Score    float64                `json:"score"`
	Type     string                 `json:"type"`
	Title    string                 `json:"title"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
	Snippet  string                 `json:"snippet"`
}

type SearchQuery struct {
	Query    string   `json:"query"`
	Type     string   `json:"type,omitempty"`     // "resource" or "prompt"
	Tags     []string `json:"tags,omitempty"`
	Limit    int      `json:"limit,omitempty"`
	Offset   int      `json:"offset,omitempty"`
	Semantic bool     `json:"semantic,omitempty"` // For future semantic search
}

func NewBleveSearch(indexPath string) (*BleveSearch, error) {
	// Create a mapping for our documents
	mapping := createIndexMapping()
	
	// Try to open existing index, create if it doesn't exist
	index, err := bleve.Open(indexPath)
	if err != nil {
		// Create new index if it doesn't exist
		index, err = bleve.New(indexPath, mapping)
		if err != nil {
			return nil, fmt.Errorf("failed to create search index: %w", err)
		}
	}

	return &BleveSearch{index: index}, nil
}

func createIndexMapping() mapping.IndexMapping {
	// Create a generic mapping
	indexMapping := bleve.NewIndexMapping()
	
	// Create mappings for different document types
	resourceMapping := bleve.NewDocumentMapping()
	
	// Text fields with different analyzers
	titleFieldMapping := bleve.NewTextFieldMapping()
	titleFieldMapping.Analyzer = "keyword"
	titleFieldMapping.Store = true
	titleFieldMapping.Index = true
	
	contentFieldMapping := bleve.NewTextFieldMapping()
	contentFieldMapping.Analyzer = "standard"
	contentFieldMapping.Store = true
	contentFieldMapping.Index = true
	
	tagsFieldMapping := bleve.NewTextFieldMapping()
	tagsFieldMapping.Analyzer = "standard"
	tagsFieldMapping.Store = true
	tagsFieldMapping.Index = true
	
	typeFieldMapping := bleve.NewTextFieldMapping()
	typeFieldMapping.Analyzer = "keyword"
	typeFieldMapping.Store = true
	typeFieldMapping.Index = true
	
	// Add field mappings
	resourceMapping.AddFieldMappingsAt("title", titleFieldMapping)
	resourceMapping.AddFieldMappingsAt("content", contentFieldMapping)
	resourceMapping.AddFieldMappingsAt("tags", tagsFieldMapping)
	resourceMapping.AddFieldMappingsAt("type", typeFieldMapping)
	resourceMapping.AddFieldMappingsAt("search_terms", contentFieldMapping)
	
	// Add document mapping
	indexMapping.AddDocumentMapping("_default", resourceMapping)
	
	return indexMapping
}

func (s *BleveSearch) Close() error {
	return s.index.Close()
}

func (s *BleveSearch) IndexResource(ctx context.Context, resource *storage.Resource) error {
	// Create a document for indexing
	doc := map[string]interface{}{
		"id":           resource.ID,
		"type":         "resource",
		"title":        resource.Title,
		"content":      resource.Content,
		"tags":         strings.Join(resource.Tags, " "),
		"search_terms": strings.Join(resource.SearchTerms, " "),
		"metadata":     resource.Metadata,
	}

	return s.index.Index(resource.ID, doc)
}

func (s *BleveSearch) IndexPrompt(ctx context.Context, prompt *storage.Prompt) error {
	// Create a document for indexing  
	doc := map[string]interface{}{
		"id":          prompt.ID,
		"type":        "prompt",
		"title":       prompt.Name,
		"content":     prompt.Description + " " + prompt.Template,
		"tags":        strings.Join(prompt.Tags, " "),
		"name":        prompt.Name,
		"description": prompt.Description,
		"template":    prompt.Template,
	}

	return s.index.Index(prompt.ID, doc)
}

func (s *BleveSearch) RemoveFromIndex(ctx context.Context, id string, docType string) error {
	return s.index.Delete(id)
}

func (s *BleveSearch) Search(ctx context.Context, query *SearchQuery) ([]*SearchResult, error) {
	if query.Limit == 0 {
		query.Limit = 10
	}

	// Build the search query
	var searchQuery blevequery.Query

	// If we have a specific query string, use it
	if query.Query != "" {
		// Create a multi-field query that searches title, content, and tags
		titleQuery := bleve.NewMatchQuery(query.Query)
		titleQuery.SetField("title")
		titleQuery.SetBoost(3.0) // Boost title matches

		contentQuery := bleve.NewMatchQuery(query.Query)
		contentQuery.SetField("content")
		contentQuery.SetBoost(1.0)

		tagsQuery := bleve.NewMatchQuery(query.Query)
		tagsQuery.SetField("tags")
		tagsQuery.SetBoost(2.0) // Boost tag matches

		searchTermsQuery := bleve.NewMatchQuery(query.Query)
		searchTermsQuery.SetField("search_terms")
		searchTermsQuery.SetBoost(2.5)

		// Combine queries with OR
		disjunctionQuery := bleve.NewDisjunctionQuery(titleQuery, contentQuery, tagsQuery, searchTermsQuery)
		searchQuery = disjunctionQuery
	} else {
		// If no query, match all documents
		searchQuery = bleve.NewMatchAllQuery()
	}

	// Add type filter if specified
	if query.Type != "" {
		typeQuery := bleve.NewTermQuery(query.Type)
		typeQuery.SetField("type")
		
		if query.Query != "" {
			// Combine with existing query using AND
			conjunctionQuery := bleve.NewConjunctionQuery(searchQuery, typeQuery)
			searchQuery = conjunctionQuery
		} else {
			searchQuery = typeQuery
		}
	}

	// Add tags filter if specified
	if len(query.Tags) > 0 {
		var tagQueries []blevequery.Query
		for _, tag := range query.Tags {
			tagQuery := bleve.NewMatchQuery(tag)
			tagQuery.SetField("tags")
			tagQueries = append(tagQueries, tagQuery)
		}
		
		tagsDisjunctionQuery := bleve.NewDisjunctionQuery(tagQueries...)
		
		if query.Query != "" || query.Type != "" {
			// Combine with existing query using AND
			conjunctionQuery := bleve.NewConjunctionQuery(searchQuery, tagsDisjunctionQuery)
			searchQuery = conjunctionQuery
		} else {
			searchQuery = tagsDisjunctionQuery
		}
	}

	// Create search request
	searchRequest := bleve.NewSearchRequest(searchQuery)
	searchRequest.Size = query.Limit
	searchRequest.From = query.Offset
	searchRequest.Highlight = bleve.NewHighlight()
	searchRequest.Fields = []string{"*"}

	// Execute search
	searchResult, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Convert results
	var results []*SearchResult
	for _, hit := range searchResult.Hits {
		result := &SearchResult{
			ID:    hit.ID,
			Score: hit.Score,
		}

		// Extract fields from the hit
		if title, ok := hit.Fields["title"].(string); ok {
			result.Title = title
		}
		if content, ok := hit.Fields["content"].(string); ok {
			result.Content = content
			// Create snippet (first 200 characters)
			if len(content) > 200 {
				result.Snippet = content[:200] + "..."
			} else {
				result.Snippet = content
			}
		}
		if docType, ok := hit.Fields["type"].(string); ok {
			result.Type = docType
		}
		if metadata, ok := hit.Fields["metadata"]; ok {
			if metadataMap, ok := metadata.(map[string]interface{}); ok {
				result.Metadata = metadataMap
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// OptimizedSearch provides a token-efficient search specifically designed for AI systems
func (s *BleveSearch) OptimizedSearch(ctx context.Context, query string, limit int) (string, error) {
	searchQuery := &SearchQuery{
		Query: query,
		Limit: limit,
	}

	results, err := s.Search(ctx, searchQuery)
	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		return "No results found", nil
	}

	// Create a compact, token-efficient response
	var output strings.Builder
	output.WriteString(fmt.Sprintf("Found %d results:\n\n", len(results)))

	for i, result := range results {
		output.WriteString(fmt.Sprintf("%d. [%s] %s (ID: %s, Score: %.2f)\n", 
			i+1, result.Type, result.Title, result.ID, result.Score))
		
		if result.Snippet != "" {
			output.WriteString(fmt.Sprintf("   %s\n", result.Snippet))
		}
		
		if i < len(results)-1 {
			output.WriteString("\n")
		}
	}

	return output.String(), nil
}