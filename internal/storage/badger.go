package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
)

type BadgerStore struct {
	db *badger.DB
}

type Resource struct {
	ID          string                 `json:"id"`
	Domain      string                 `json:"domain,omitempty"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Metadata    map[string]interface{} `json:"metadata"`
	Tags        []string               `json:"tags"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	SearchTerms []string               `json:"search_terms"`
}

type Prompt struct {
	ID          string            `json:"id"`
	Domain      string            `json:"domain,omitempty"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Template    string            `json:"template"`
	Variables   map[string]string `json:"variables"`
	Tags        []string          `json:"tags"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

func NewBadgerStore(path string) (*BadgerStore, error) {
	opts := badger.DefaultOptions(path)
	opts.Logger = nil // Disable badger logs
	
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger database: %w", err)
	}

	return &BadgerStore{db: db}, nil
}

func (s *BadgerStore) Close() error {
	return s.db.Close()
}

// Domain-scoped resource operations
func (s *BadgerStore) StoreResource(ctx context.Context, resource *Resource) error {
	if resource == nil {
		return fmt.Errorf("resource cannot be nil")
	}
	return s.StoreResourceInDomain(ctx, resource, resource.Domain)
}

func (s *BadgerStore) StoreResourceInDomain(ctx context.Context, resource *Resource, domain string) error {
	if resource == nil {
		return fmt.Errorf("resource cannot be nil")
	}
	if resource.ID == "" {
		return fmt.Errorf("resource ID cannot be empty")
	}
	
	resource.UpdatedAt = time.Now()
	if resource.CreatedAt.IsZero() {
		resource.CreatedAt = resource.UpdatedAt
	}
	
	// Set domain if specified
	if domain != "" {
		resource.Domain = domain
	}

	data, err := json.Marshal(resource)
	if err != nil {
		return fmt.Errorf("failed to marshal resource: %w", err)
	}

	return s.db.Update(func(txn *badger.Txn) error {
		key := s.buildResourceKey(resource.ID, domain)
		return txn.Set([]byte(key), data)
	})
}

func (s *BadgerStore) GetResource(ctx context.Context, id string) (*Resource, error) {
	if id == "" {
		return nil, fmt.Errorf("resource ID cannot be empty")
	}
	
	var resource Resource
	
	err := s.db.View(func(txn *badger.Txn) error {
		// Use domain-aware key building to match how resources are stored
		key := s.buildResourceKey(id, "default")
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
			return nil, fmt.Errorf("resource not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get resource: %w", err)
	}

	return &resource, nil
}

func (s *BadgerStore) ListResources(ctx context.Context, limit int, offset int) ([]*Resource, error) {
	var resources []*Resource
	
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = limit
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("resource:")
		count := 0
		skipped := 0

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			if skipped < offset {
				skipped++
				continue
			}

			if count >= limit {
				break
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
		return nil
	})

	return resources, err
}

func (s *BadgerStore) DeleteResource(ctx context.Context, id string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		key := s.buildResourceKey(id, "default")
		return txn.Delete([]byte(key))
	})
}

// Prompt operations
func (s *BadgerStore) StorePrompt(ctx context.Context, prompt *Prompt) error {
	if prompt == nil {
		return fmt.Errorf("prompt cannot be nil")
	}
	if prompt.ID == "" {
		return fmt.Errorf("prompt ID cannot be empty")
	}
	
	prompt.UpdatedAt = time.Now()
	if prompt.CreatedAt.IsZero() {
		prompt.CreatedAt = prompt.UpdatedAt
	}

	data, err := json.Marshal(prompt)
	if err != nil {
		return fmt.Errorf("failed to marshal prompt: %w", err)
	}

	return s.db.Update(func(txn *badger.Txn) error {
		key := fmt.Sprintf("prompt:%s", prompt.ID)
		return txn.Set([]byte(key), data)
	})
}

func (s *BadgerStore) GetPrompt(ctx context.Context, id string) (*Prompt, error) {
	var prompt Prompt
	
	err := s.db.View(func(txn *badger.Txn) error {
		key := fmt.Sprintf("prompt:%s", id)
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
			return nil, fmt.Errorf("prompt not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get prompt: %w", err)
	}

	return &prompt, nil
}

func (s *BadgerStore) ListPrompts(ctx context.Context, limit int, offset int) ([]*Prompt, error) {
	var prompts []*Prompt
	
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = limit
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("prompt:")
		count := 0
		skipped := 0

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			if skipped < offset {
				skipped++
				continue
			}

			if count >= limit {
				break
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
		return nil
	})

	return prompts, err
}

func (s *BadgerStore) DeletePrompt(ctx context.Context, id string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		key := fmt.Sprintf("prompt:%s", id)
		return txn.Delete([]byte(key))
	})
}

// Search helper methods
func (s *BadgerStore) SearchResourcesByTags(ctx context.Context, tags []string, limit int) ([]*Resource, error) {
	var resources []*Resource
	
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("resource:")
		count := 0

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
		return nil
	})

	return resources, err
}