package domain

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Domain represents a namespace for organizing resources and prompts
type Domain struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Parent      string            `json:"parent,omitempty"`      // Parent domain for hierarchy
	Path        string            `json:"path"`                  // Full hierarchical path
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Active      bool              `json:"active"`
}

// DomainConfig represents domain-specific configuration
type DomainConfig struct {
	DefaultDomain   string            `json:"default_domain"`
	IsolationMode   string            `json:"isolation_mode"` // "strict", "hierarchical", "shared"
	AllowCrossDomain bool             `json:"allow_cross_domain"`
	DomainSettings  map[string]string `json:"domain_settings,omitempty"`
}

// DomainManager handles domain operations and scoping
type DomainManager struct {
	domains map[string]*Domain
	config  *DomainConfig
}

// DomainScope represents the current domain context for operations
type DomainScope struct {
	Current    string   `json:"current"`           // Current active domain
	Ancestry   []string `json:"ancestry"`          // Parent domains in order
	Children   []string `json:"children"`          // Direct child domains
	Searchable []string `json:"searchable"`        // Domains that can be searched
}

const (
	// Domain naming constraints
	DomainNamePattern = `^[a-z0-9][a-z0-9\-_]*[a-z0-9]$`
	DomainSeparator   = "/"
	DefaultDomain     = "default"
	
	// Isolation modes
	IsolationStrict       = "strict"       // Complete isolation between domains
	IsolationHierarchical = "hierarchical" // Parent can access children, children inherit from parents
	IsolationShared       = "shared"       // All domains can see each other
)

func NewDomainManager(config *DomainConfig) *DomainManager {
	if config == nil {
		config = &DomainConfig{
			DefaultDomain:   DefaultDomain,
			IsolationMode:   IsolationHierarchical,
			AllowCrossDomain: true,
		}
	}
	
	dm := &DomainManager{
		domains: make(map[string]*Domain),
		config:  config,
	}
	
	// Create default domain
	dm.createDefaultDomain()
	
	return dm
}

func (dm *DomainManager) createDefaultDomain() {
	defaultDomain := &Domain{
		ID:          DefaultDomain,
		Name:        "Default Domain",
		Description: "Default domain for general resources",
		Path:        DomainSeparator + DefaultDomain,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Active:      true,
	}
	dm.domains[DefaultDomain] = defaultDomain
}

// CreateDomain creates a new domain with optional parent
func (dm *DomainManager) CreateDomain(ctx context.Context, id, name, description, parent string) (*Domain, error) {
	// Validate domain ID
	if !dm.isValidDomainName(id) {
		return nil, fmt.Errorf("invalid domain ID: must match pattern %s", DomainNamePattern)
	}
	
	// Check if domain already exists
	if _, exists := dm.domains[id]; exists {
		return nil, fmt.Errorf("domain already exists: %s", id)
	}
	
	// Validate parent if specified
	var path string
	if parent != "" {
		parentDomain, exists := dm.domains[parent]
		if !exists {
			return nil, fmt.Errorf("parent domain not found: %s", parent)
		}
		path = parentDomain.Path + DomainSeparator + id
	} else {
		path = DomainSeparator + id
	}
	
	domain := &Domain{
		ID:          id,
		Name:        name,
		Description: description,
		Parent:      parent,
		Path:        path,
		Metadata:    make(map[string]string),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Active:      true,
	}
	
	dm.domains[id] = domain
	return domain, nil
}

// GetDomain retrieves a domain by ID
func (dm *DomainManager) GetDomain(id string) (*Domain, error) {
	domain, exists := dm.domains[id]
	if !exists {
		return nil, fmt.Errorf("domain not found: %s", id)
	}
	return domain, nil
}

// ListDomains returns all domains, optionally filtered by parent
func (dm *DomainManager) ListDomains(parent string) []*Domain {
	var domains []*Domain
	for _, domain := range dm.domains {
		if parent == "" || domain.Parent == parent {
			domains = append(domains, domain)
		}
	}
	return domains
}

// GetDomainScope calculates the scope for a given domain
func (dm *DomainManager) GetDomainScope(domainID string) (*DomainScope, error) {
	domain, exists := dm.domains[domainID]
	if !exists {
		return nil, fmt.Errorf("domain not found: %s", domainID)
	}
	
	scope := &DomainScope{
		Current:    domainID,
		Ancestry:   dm.getAncestry(domain),
		Children:   dm.getChildren(domainID),
		Searchable: dm.getSearchableDomains(domainID),
	}
	
	return scope, nil
}

// getAncestry returns the parent hierarchy for a domain
func (dm *DomainManager) getAncestry(domain *Domain) []string {
	var ancestry []string
	visited := make(map[string]bool)
	current := domain
	maxDepth := 50 // Prevent infinite loops
	depth := 0
	
	for current.Parent != "" && depth < maxDepth {
		// Check for circular dependency
		if visited[current.Parent] {
			break
		}
		
		ancestry = append([]string{current.Parent}, ancestry...)
		visited[current.Parent] = true
		
		if parent, exists := dm.domains[current.Parent]; exists {
			current = parent
			depth++
		} else {
			break
		}
	}
	
	return ancestry
}

// getChildren returns direct children of a domain
func (dm *DomainManager) getChildren(domainID string) []string {
	var children []string
	for _, domain := range dm.domains {
		if domain.Parent == domainID {
			children = append(children, domain.ID)
		}
	}
	return children
}

// getSearchableDomains returns domains that can be searched based on isolation mode
func (dm *DomainManager) getSearchableDomains(domainID string) []string {
	switch dm.config.IsolationMode {
	case IsolationStrict:
		return []string{domainID}
		
	case IsolationShared:
		var all []string
		for id := range dm.domains {
			all = append(all, id)
		}
		return all
		
	case IsolationHierarchical:
		fallthrough
	default:
		return dm.getHierarchicalSearchable(domainID)
	}
}

// getHierarchicalSearchable returns searchable domains in hierarchical mode
func (dm *DomainManager) getHierarchicalSearchable(domainID string) []string {
	domain, exists := dm.domains[domainID]
	if !exists {
		return []string{domainID}
	}
	
	searchable := []string{domainID}
	
	// Add ancestry (can search parent domains) - use direct getAncestry to avoid circular dependency
	ancestry := dm.getAncestry(domain)
	searchable = append(searchable, ancestry...)
	
	// Add children (can search child domains)
	children := dm.getChildren(domainID)
	searchable = append(searchable, children...)
	
	// Add descendants recursively
	for _, child := range children {
		descendants := dm.getDescendants(child)
		searchable = append(searchable, descendants...)
	}
	
	return dm.uniqueStrings(searchable)
}

// getDescendants recursively gets all descendant domains
func (dm *DomainManager) getDescendants(domainID string) []string {
	return dm.getDescendantsWithVisited(domainID, make(map[string]bool), 0)
}

// getDescendantsWithVisited is a helper that tracks visited domains to prevent infinite recursion
func (dm *DomainManager) getDescendantsWithVisited(domainID string, visited map[string]bool, depth int) []string {
	var descendants []string
	maxDepth := 50 // Prevent infinite loops
	
	if visited[domainID] || depth > maxDepth {
		return descendants
	}
	
	visited[domainID] = true
	children := dm.getChildren(domainID)
	
	for _, child := range children {
		descendants = append(descendants, child)
		descendants = append(descendants, dm.getDescendantsWithVisited(child, visited, depth+1)...)
	}
	
	return descendants
}

// ArchiveDomain marks a domain as inactive
func (dm *DomainManager) ArchiveDomain(ctx context.Context, domainID string) error {
	if domainID == DefaultDomain {
		return fmt.Errorf("cannot archive default domain")
	}
	
	domain, exists := dm.domains[domainID]
	if !exists {
		return fmt.Errorf("domain not found: %s", domainID)
	}
	
	domain.Active = false
	domain.UpdatedAt = time.Now()
	
	return nil
}

// DeleteDomain removes a domain and all its children
func (dm *DomainManager) DeleteDomain(ctx context.Context, domainID string, force bool) error {
	if domainID == DefaultDomain {
		return fmt.Errorf("cannot delete default domain")
	}
	
	// Check for children
	children := dm.getChildren(domainID)
	if len(children) > 0 && !force {
		return fmt.Errorf("domain has children, use force=true to delete: %v", children)
	}
	
	// Delete children recursively if force is true
	if force {
		for _, child := range children {
			dm.DeleteDomain(ctx, child, true)
		}
	}
	
	delete(dm.domains, domainID)
	return nil
}

// ValidateDomainAccess checks if a domain can access another domain
func (dm *DomainManager) ValidateDomainAccess(sourceDomain, targetDomain string) bool {
	if sourceDomain == targetDomain {
		return true
	}
	
	scope, err := dm.GetDomainScope(sourceDomain)
	if err != nil {
		return false
	}
	
	for _, searchable := range scope.Searchable {
		if searchable == targetDomain {
			return true
		}
	}
	
	return false
}

// ResolveDomainPath resolves a hierarchical domain path to domain ID
func (dm *DomainManager) ResolveDomainPath(path string) (string, error) {
	// Handle direct domain ID
	if _, exists := dm.domains[path]; exists {
		return path, nil
	}
	
	// Handle hierarchical path
	parts := strings.Split(path, DomainSeparator)
	if len(parts) == 1 {
		return "", fmt.Errorf("domain not found: %s", path)
	}
	
	// Find domain with matching path
	for id, domain := range dm.domains {
		if domain.Path == path {
			return id, nil
		}
	}
	
	return "", fmt.Errorf("domain path not found: %s", path)
}

// GetDomainPrefix returns the storage/index prefix for a domain
func (dm *DomainManager) GetDomainPrefix(domainID string) string {
	if domainID == DefaultDomain {
		return ""
	}
	return fmt.Sprintf("domain:%s:", domainID)
}

// ParseResourceID extracts domain and resource ID from a domain-prefixed ID with shorthand support
// Examples:
//   - "::abc123" -> domain="default", id="abc123"
//   - "domain:proj1:abc123" -> domain="proj1", id="abc123"
//   - "proj1:abc123" -> domain="proj1", id="abc123"
//   - "abc123" -> domain="default", id="abc123"
func (dm *DomainManager) ParseResourceID(resourceID string) (domain, id string) {
	// Shorthand notation for default domain
	if strings.HasPrefix(resourceID, "::") {
		return DefaultDomain, strings.TrimPrefix(resourceID, "::")
	}
	
	// Explicit domain prefix (legacy support)
	if strings.HasPrefix(resourceID, "domain:") {
		parts := strings.SplitN(resourceID, ":", 3)
		if len(parts) == 3 {
			return parts[1], parts[2]
		}
	}
	
	// Domain:ID format
	if strings.Contains(resourceID, ":") && !strings.HasPrefix(resourceID, "::") {
		parts := strings.SplitN(resourceID, ":", 2)
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
	}
	
	// Default domain
	return DefaultDomain, resourceID
}

// BuildResourceID creates a domain-prefixed resource ID with shorthand notation
func (dm *DomainManager) BuildResourceID(domainID, resourceID string) string {
	if domainID == DefaultDomain || domainID == "" {
		return fmt.Sprintf("::%s", resourceID) // Shorthand notation
	}
	return fmt.Sprintf("%s:%s", domainID, resourceID)
}

// NormalizeResourceID converts various input formats to a canonical format
// Examples:
//   - "::abc123" -> "::abc123" (already canonical)
//   - "domain:proj1:abc123" -> "proj1:abc123" (convert legacy format)
//   - "abc123" -> "::abc123" (add default domain prefix)
func (dm *DomainManager) NormalizeResourceID(input string) string {
	domain, id := dm.ParseResourceID(input)
	return dm.BuildResourceID(domain, id)
}

// Helper functions

func (dm *DomainManager) isValidDomainName(name string) bool {
	matched, _ := regexp.MatchString(DomainNamePattern, name)
	return matched && len(name) <= 64
}

func (dm *DomainManager) uniqueStrings(slice []string) []string {
	seen := make(map[string]bool)
	var result []string
	
	for _, str := range slice {
		if !seen[str] {
			seen[str] = true
			result = append(result, str)
		}
	}
	
	return result
}

// GetConfig returns the current domain configuration
func (dm *DomainManager) GetConfig() *DomainConfig {
	return dm.config
}

// UpdateConfig updates the domain configuration
func (dm *DomainManager) UpdateConfig(config *DomainConfig) {
	dm.config = config
}

// GetCurrentDomain returns the default domain or specified domain
func (dm *DomainManager) GetCurrentDomain() string {
	return dm.config.DefaultDomain
}

// SetCurrentDomain sets the default domain
func (dm *DomainManager) SetCurrentDomain(domainID string) error {
	if _, exists := dm.domains[domainID]; !exists {
		return fmt.Errorf("domain not found: %s", domainID)
	}
	dm.config.DefaultDomain = domainID
	return nil
}