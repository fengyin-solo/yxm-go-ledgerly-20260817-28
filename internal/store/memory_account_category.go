package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/example/ledgerly/internal/logger"
	"github.com/example/ledgerly/internal/model"
)

// MemoryAccountStore is an in-memory account store with JSON persistence.
type MemoryAccountStore struct {
	mu    sync.RWMutex
	items map[string]*model.Account
	path  string
	dirty bool
	log   *logger.Logger
	stop  chan struct{}
	done  chan struct{}
}

// NewMemoryAccountStore creates a MemoryAccountStore.
func NewMemoryAccountStore(path string, log *logger.Logger, interval time.Duration) (*MemoryAccountStore, error) {
	if log == nil {
		log = logger.Default()
	}
	s := &MemoryAccountStore{
		items: make(map[string]*model.Account),
		path:  path,
		log:   log,
		stop:  make(chan struct{}),
		done:  make(chan struct{}),
	}
	if path != "" {
		if err := s.load(); err != nil {
			log.Warn("account store load failed", "err", err.Error())
		}
		go s.saveLoop(interval)
	} else {
		close(s.done)
	}
	return s, nil
}

func (s *MemoryAccountStore) Create(ctx context.Context, a *model.Account) error {
	if a == nil || a.ID == "" {
		return model.ErrInvalidInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.items[a.ID]; exists {
		return model.ErrAlreadyExists
	}
	s.items[a.ID] = a.Clone()
	s.dirty = true
	return nil
}

func (s *MemoryAccountStore) GetByID(ctx context.Context, id string) (*model.Account, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.items[id]
	if !ok {
		return nil, model.ErrNotFound
	}
	return a.Clone(), nil
}

func (s *MemoryAccountStore) Update(ctx context.Context, a *model.Account) error {
	if a == nil || a.ID == "" {
		return model.ErrInvalidInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[a.ID]; !ok {
		return model.ErrNotFound
	}
	s.items[a.ID] = a.Clone()
	s.dirty = true
	return nil
}

func (s *MemoryAccountStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return model.ErrNotFound
	}
	delete(s.items, id)
	s.dirty = true
	return nil
}

func (s *MemoryAccountStore) List(ctx context.Context, f model.AccountFilter) ([]*model.Account, error) {
	s.mu.RLock()
	var out []*model.Account
	for _, a := range s.items {
		if f.Matches(a) {
			out = append(out, a.Clone())
		}
	}
	s.mu.RUnlock()
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (s *MemoryAccountStore) Close() error {
	if s.path == "" {
		return nil
	}
	s.mu.Lock()
	close(s.stop)
	s.mu.Unlock()
	<-s.done
	return s.flush()
}

func (s *MemoryAccountStore) saveLoop(interval time.Duration) {
	defer close(s.done)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			s.mu.RLock()
			dirty := s.dirty
			s.mu.RUnlock()
			if dirty {
				s.flush()
			}
		}
	}
}

func (s *MemoryAccountStore) flush() error {
	if s.path == "" {
		return nil
	}
	s.mu.Lock()
	snapshot := make([]*model.Account, 0, len(s.items))
	for _, a := range s.items {
		snapshot = append(snapshot, a.Clone())
	}
	s.dirty = false
	s.mu.Unlock()
	sort.Slice(snapshot, func(i, j int) bool { return snapshot[i].Name < snapshot[j].Name })
	payload, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal accounts: %w", err)
	}
	dir := filepath.Dir(s.path)
	if dir != "" && dir != "." {
		os.MkdirAll(dir, 0o755) // #nosec G301
	}
	tmp, err := os.CreateTemp(dir, ".ledgerly-accounts-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(payload); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, s.path)
}

func (s *MemoryAccountStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if len(data) == 0 {
		return nil
	}
	var items []*model.Account
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("corrupt account data: %w", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, a := range items {
		if a != nil && a.ID != "" {
			s.items[a.ID] = a
		}
	}
	return nil
}

// MemoryCategoryStore is an in-memory category store with JSON persistence.
type MemoryCategoryStore struct {
	mu    sync.RWMutex
	items map[string]*model.Category
	path  string
	dirty bool
	log   *logger.Logger
	stop  chan struct{}
	done  chan struct{}
}

// NewMemoryCategoryStore creates a MemoryCategoryStore.
func NewMemoryCategoryStore(path string, log *logger.Logger, interval time.Duration) (*MemoryCategoryStore, error) {
	if log == nil {
		log = logger.Default()
	}
	s := &MemoryCategoryStore{
		items: make(map[string]*model.Category),
		path:  path,
		log:   log,
		stop:  make(chan struct{}),
		done:  make(chan struct{}),
	}
	if path != "" {
		if err := s.load(); err != nil {
			log.Warn("category store load failed", "err", err.Error())
		}
		go s.saveLoop(interval)
	} else {
		close(s.done)
	}
	return s, nil
}

func (s *MemoryCategoryStore) Create(ctx context.Context, c *model.Category) error {
	if c == nil || c.ID == "" {
		return model.ErrInvalidInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.items[c.ID]; exists {
		return model.ErrAlreadyExists
	}
	s.items[c.ID] = c.Clone()
	s.dirty = true
	return nil
}

func (s *MemoryCategoryStore) GetByID(ctx context.Context, id string) (*model.Category, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.items[id]
	if !ok {
		return nil, model.ErrNotFound
	}
	return c, nil
}

func (s *MemoryCategoryStore) Update(ctx context.Context, c *model.Category) error {
	if c == nil || c.ID == "" {
		return model.ErrInvalidInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[c.ID]; !ok {
		return model.ErrNotFound
	}
	s.items[c.ID] = c.Clone()
	s.dirty = true
	return nil
}

func (s *MemoryCategoryStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return model.ErrNotFound
	}
	delete(s.items, id)
	s.dirty = true
	return nil
}

func (s *MemoryCategoryStore) List(ctx context.Context, f model.CategoryFilter) ([]*model.Category, error) {
	s.mu.RLock()
	var out []*model.Category
	for _, c := range s.items {
		if f.Matches(c) {
			out = append(out, c.Clone())
		}
	}
	s.mu.RUnlock()
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (s *MemoryCategoryStore) Close() error {
	if s.path == "" {
		return nil
	}
	s.mu.Lock()
	close(s.stop)
	s.mu.Unlock()
	<-s.done
	return s.flush()
}

func (s *MemoryCategoryStore) saveLoop(interval time.Duration) {
	defer close(s.done)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			s.mu.RLock()
			dirty := s.dirty
			s.mu.RUnlock()
			if dirty {
				s.flush()
			}
		}
	}
}

func (s *MemoryCategoryStore) flush() error {
	if s.path == "" {
		return nil
	}
	s.mu.Lock()
	snapshot := make([]*model.Category, 0, len(s.items))
	for _, c := range s.items {
		snapshot = append(snapshot, c.Clone())
	}
	s.dirty = false
	s.mu.Unlock()
	sort.Slice(snapshot, func(i, j int) bool { return snapshot[i].Name < snapshot[j].Name })
	payload, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal categories: %w", err)
	}
	dir := filepath.Dir(s.path)
	if dir != "" && dir != "." {
		os.MkdirAll(dir, 0o755) // #nosec G301
	}
	tmp, err := os.CreateTemp(dir, ".ledgerly-categories-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(payload); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, s.path)
}

func (s *MemoryCategoryStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if len(data) == 0 {
		return nil
	}
	var items []*model.Category
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("corrupt category data: %w", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, c := range items {
		if c != nil && c.ID != "" {
			s.items[c.ID] = c
		}
	}
	return nil
}
