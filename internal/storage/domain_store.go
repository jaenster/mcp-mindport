package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
)

// Domain-aware storage methods with shorthand notation support

// parseDomainKey parses a key with optional shorthand notation (::)
// Examples:
//   - "resource::abc123" -> domain="default", resourceID="abc123"
//   - "resource:domain:proj1:abc123" -> domain="proj1", resourceID="abc123"
//   - "resource:abc123" -> domain="default", resourceID="abc123"
func (s *BadgerStore) parseDomainKey(key string) (itemType, domain, itemID string, err error) {
	parts := strings.Split(key, ":")
	
	if len(parts) < 2 {
		return "", "", "", fmt.Errorf("invalid key format: %s", key)
	}
	
	itemType = parts[0]
	
	// Handle shorthand notation: resource::itemID
	if len(parts) == 3 && parts[1] == "" {
		return itemType, "default", parts[2], nil
	}
	
	// Handle explicit domain: resource:domain:domainName:itemID
	if len(parts) == 4 && parts[1] == "domain" {
		return itemType, parts[2], parts[3], nil
	}
	
	// Handle default domain: resource:itemID
	if len(parts) == 2 {
		return itemType, "default", parts[1], nil
	}
	
	return "", "", "", fmt.Errorf("unsupported key format: %s", key)
}

// buildResourceKey creates a domain-scoped key for resources with shorthand notation support
func (s *BadgerStore) buildResourceKey(resourceID, domain string) string {
	if domain == "" || domain == "default" {
		return fmt.Sprintf("resource::%s", resourceID) // Shorthand notation
	}
	return fmt.Sprintf("resource:domain:%s:%s", domain, resourceID)
}

// buildPromptKey creates a domain-scoped key for prompts with shorthand notation support
func (s *BadgerStore) buildPromptKey(promptID, domain string) string {
	if domain == "" || domain == "default" {
		return fmt.Sprintf("prompt::%s", promptID) // Shorthand notation
	}
	return fmt.Sprintf("prompt:domain:%s:%s", domain, promptID)
}

// getDomainPrefix returns the prefix for scanning domain-scoped items with shorthand support
func (s *BadgerStore) getDomainPrefix(itemType, domain string) string {
	if domain == "" || domain == "default" {
		return fmt.Sprintf("%s::", itemType) // Shorthand notation prefix
	}
	return fmt.Sprintf("%s:domain:%s:", itemType, domain)
}

// buildResourceKeyFromUserInput builds a resource key from user input, supporting shorthand notation
// Examples:
//   - "::abc123" -> "resource::abc123" (default domain)
//   - "proj1:abc123" -> "resource:domain:proj1:abc123" (explicit domain)
//   - "abc123" -> "resource::abc123" (default domain, no prefix)
func (s *BadgerStore) buildResourceKeyFromUserInput(input string) string {
	if strings.HasPrefix(input, "::") {
		// User provided ::resourceID
		return fmt.Sprintf("resource%s", input)
	}
	
	if strings.Contains(input, ":") && !strings.HasPrefix(input, "::") {
		// User provided domain:resourceID
		parts := strings.SplitN(input, ":", 2)
		if len(parts) == 2 {
			return s.buildResourceKey(parts[1], parts[0])
		}
	}
	
	// User provided just resourceID, use default domain
	return s.buildResourceKey(input, "default")
}

// buildPromptKeyFromUserInput builds a prompt key from user input, supporting shorthand notation
func (s *BadgerStore) buildPromptKeyFromUserInput(input string) string {
	if strings.HasPrefix(input, "::") {
		// User provided ::promptID
		return fmt.Sprintf("prompt%s", input)
	}
	
	if strings.Contains(input, ":") && !strings.HasPrefix(input, "::") {
		// User provided domain:promptID
		parts := strings.SplitN(input, ":", 2)
		if len(parts) == 2 {
			return s.buildPromptKey(parts[1], parts[0])
		}
	}
	
	// User provided just promptID, use default domain
	return s.buildPromptKey(input, "default")
}

// GetResourceInDomain retrieves a resource from a specific domain
func (s *BadgerStore) GetResourceInDomain(ctx context.Context, id, domain string) (*Resource, error) {
	var resource Resource
	
	err := s.db.View(func(txn *badger.Txn) error {
		key := s.buildResourceKey(id, domain)
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &resource)
		})
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, fmt.Errorf("resource not found: %s in domain %s", id, domain)
		}
		return nil, fmt.Errorf("failed to get resource: %w", err)
	}

	return &resource, nil
}

// ListResourcesInDomain returns resources from specific domain(s)
func (s *BadgerStore) ListResourcesInDomain(ctx context.Context, domains []string, limit int, offset int) ([]*Resource, error) {
	var resources []*Resource
	
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = limit
		it := txn.NewIterator(opts)
		defer it.Close()

		count := 0
		skipped := 0

		// If no domains specified, search all
		if len(domains) == 0 || (len(domains) == 1 && domains[0] == "") {
			domains = []string{"", "default"}
		}

		for _, domain := range domains {
			prefix := []byte(s.getDomainPrefix("resource", domain))
			
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				if skipped < offset {
					skipped++
					continue
				}

				if count >= limit {
					return nil // Break out of all loops
				}

				item := it.Item()
				err := item.Value(func(val []byte) error {
					var resource Resource
					if err := json.Unmarshal(val, &resource); err != nil {
						return err
					}
					resources = append(resources, &resource)
					return nil
				})

				if err != nil {
					return err
				}
				count++
			}
			
			if count >= limit {
				break
			}
		}
		return nil
	})

	return resources, err
}

// DeleteResourceInDomain removes a resource from a specific domain
func (s *BadgerStore) DeleteResourceInDomain(ctx context.Context, id, domain string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		key := s.buildResourceKey(id, domain)
		return txn.Delete([]byte(key))
	})
}

// StorePromptInDomain stores a prompt in a specific domain
func (s *BadgerStore) StorePromptInDomain(ctx context.Context, prompt *Prompt, domain string) error {
	prompt.UpdatedAt = time.Now()
	if prompt.CreatedAt.IsZero() {
		prompt.CreatedAt = prompt.UpdatedAt
	}
	
	// Set domain if specified
	if domain != "" {
		prompt.Domain = domain
	}

	data, err := json.Marshal(prompt)
	if err != nil {
		return fmt.Errorf("failed to marshal prompt: %w", err)
	}

	return s.db.Update(func(txn *badger.Txn) error {
		key := s.buildPromptKey(prompt.ID, domain)
		return txn.Set([]byte(key), data)
	})
}

// GetPromptInDomain retrieves a prompt from a specific domain
func (s *BadgerStore) GetPromptInDomain(ctx context.Context, id, domain string) (*Prompt, error) {
	var prompt Prompt
	
	err := s.db.View(func(txn *badger.Txn) error {
		key := s.buildPromptKey(id, domain)
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &prompt)
		})
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, fmt.Errorf("prompt not found: %s in domain %s", id, domain)
		}
		return nil, fmt.Errorf("failed to get prompt: %w", err)
	}

	return &prompt, nil
}

// ListPromptsInDomain returns prompts from specific domain(s)
func (s *BadgerStore) ListPromptsInDomain(ctx context.Context, domains []string, limit int, offset int) ([]*Prompt, error) {
	var prompts []*Prompt
	
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = limit
		it := txn.NewIterator(opts)
		defer it.Close()

		count := 0
		skipped := 0

		// If no domains specified, search all
		if len(domains) == 0 || (len(domains) == 1 && domains[0] == "") {
			domains = []string{"", "default"}
		}

		for _, domain := range domains {
			prefix := []byte(s.getDomainPrefix("prompt", domain))
			
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				if skipped < offset {
					skipped++
					continue
				}

				if count >= limit {
					return nil
				}

				item := it.Item()
				err := item.Value(func(val []byte) error {
					var prompt Prompt
					if err := json.Unmarshal(val, &prompt); err != nil {
						return err
					}
					prompts = append(prompts, &prompt)
					return nil
				})

				if err != nil {
					return err
				}
				count++
			}
			
			if count >= limit {
				break
			}
		}
		return nil
	})

	return prompts, err
}

// DeletePromptInDomain removes a prompt from a specific domain
func (s *BadgerStore) DeletePromptInDomain(ctx context.Context, id, domain string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		key := s.buildPromptKey(id, domain)
		return txn.Delete([]byte(key))
	})
}

// SearchResourcesByTagsInDomain searches within specific domains
func (s *BadgerStore) SearchResourcesByTagsInDomain(ctx context.Context, tags []string, domains []string, limit int) ([]*Resource, error) {
	var resources []*Resource
	
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		count := 0

		// If no domains specified, search all
		if len(domains) == 0 || (len(domains) == 1 && domains[0] == "") {
			domains = []string{"", "default"}
		}

		for _, domain := range domains {
			prefix := []byte(s.getDomainPrefix("resource", domain))
			
			for it.Seek(prefix); it.ValidForPrefix(prefix) && count < limit; it.Next() {
				item := it.Item()
				err := item.Value(func(val []byte) error {
					var resource Resource
					if err := json.Unmarshal(val, &resource); err != nil {
						return err
					}

					// Check if resource has any of the specified tags
					for _, tag := range tags {
						for _, resourceTag := range resource.Tags {
							if strings.EqualFold(tag, resourceTag) {
								resources = append(resources, &resource)
								count++
								return nil
							}
						}
					}
					return nil
				})

				if err != nil {
					return err
				}
			}
			
			if count >= limit {
				break
			}
		}
		return nil
	})

	return resources, err
}

// GetDomainStats returns statistics about a domain
func (s *BadgerStore) GetDomainStats(ctx context.Context, domain string) (map[string]int, error) {
	stats := map[string]int{
		"resources": 0,
		"prompts":   0,
	}
	
	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		// Count resources
		resourcePrefix := []byte(s.getDomainPrefix("resource", domain))
		for it.Seek(resourcePrefix); it.ValidForPrefix(resourcePrefix); it.Next() {
			stats["resources"]++
		}

		// Count prompts
		promptPrefix := []byte(s.getDomainPrefix("prompt", domain))
		for it.Seek(promptPrefix); it.ValidForPrefix(promptPrefix); it.Next() {
			stats["prompts"]++
		}

		return nil
	})

	return stats, err
}

// ListAllDomains returns all domains that have data
func (s *BadgerStore) ListAllDomains(ctx context.Context) ([]string, error) {
	domainSet := make(map[string]bool)
	
	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			key := string(it.Item().Key())
			
			// Parse domain from key
			if strings.HasPrefix(key, "resource:domain:") || strings.HasPrefix(key, "prompt:domain:") {
				parts := strings.Split(key, ":")
				if len(parts) >= 3 {
					domain := parts[2]
					domainSet[domain] = true
				}
			} else if strings.HasPrefix(key, "resource:") || strings.HasPrefix(key, "prompt:") {
				domainSet["default"] = true
			}
		}

		return nil
	})

	var domains []string
	for domain := range domainSet {
		domains = append(domains, domain)
	}

	return domains, err
}