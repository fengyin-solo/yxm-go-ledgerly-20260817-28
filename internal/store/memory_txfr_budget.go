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

// MemoryTransactionStore is an in-memory transaction store with JSON persistence.
type MemoryTransactionStore struct {
	mu    sync.RWMutex
	items map[string]*model.Transaction
	path  string
	dirty bool
	log   *logger.Logger
	stop  chan struct{}
	done  chan struct{}
}

// NewMemoryTransactionStore creates a MemoryTransactionStore.
func NewMemoryTransactionStore(path string, log *logger.Logger, interval time.Duration) (*MemoryTransactionStore, error) {
	if log == nil {
		log = logger.Default()
	}
	s := &MemoryTransactionStore{
		items: make(map[string]*model.Transaction),
		path:  path,
		log:   log,
		stop:  make(chan struct{}),
		done:  make(chan struct{}),
	}
	if path != "" {
		if err := s.load(); err != nil {
			log.Warn("transaction store load failed", "err", err.Error())
		}
		go s.saveLoop(interval)
	} else {
		close(s.done)
	}
	return s, nil
}

func (s *MemoryTransactionStore) Create(ctx context.Context, t *model.Transaction) error {
	if t == nil || t.ID == "" {
		return model.ErrInvalidInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[t.ID] = t.Clone()
	s.dirty = true
	return nil
}

func (s *MemoryTransactionStore) GetByID(ctx context.Context, id string) (*model.Transaction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.items[id]
	if !ok {
		return nil, model.ErrNotFound
	}
	return t.Clone(), nil
}

func (s *MemoryTransactionStore) List(ctx context.Context, f model.TransactionFilter) ([]*model.Transaction, error) {
	s.mu.RLock()
	var out []*model.Transaction
	for _, t := range s.items {
		if f.Matches(t) {
			out = append(out, t.Clone())
		}
	}
	s.mu.RUnlock()
	sort.Slice(out, func(i, j int) bool { return out[i].Date.After(out[j].Date) })
	return out, nil
}

func (s *MemoryTransactionStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return model.ErrNotFound
	}
	delete(s.items, id)
	s.dirty = true
	return nil
}

func (s *MemoryTransactionStore) Close() error {
	if s.path == "" {
		return nil
	}
	s.mu.Lock()
	close(s.stop)
	s.mu.Unlock()
	<-s.done
	return s.flush()
}

func (s *MemoryTransactionStore) saveLoop(interval time.Duration) {
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

func (s *MemoryTransactionStore) flush() error {
	if s.path == "" {
		return nil
	}
	s.mu.Lock()
	snapshot := make([]*model.Transaction, 0, len(s.items))
	for _, t := range s.items {
		snapshot = append(snapshot, t.Clone())
	}
	s.dirty = false
	s.mu.Unlock()
	sort.Slice(snapshot, func(i, j int) bool { return snapshot[i].Date.After(snapshot[j].Date) })
	payload, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal transactions: %w", err)
	}
	dir := filepath.Dir(s.path)
	if dir != "" && dir != "." {
		os.MkdirAll(dir, 0o755) // #nosec G301
	}
	tmp, err := os.CreateTemp(dir, ".ledgerly-transactions-*.tmp")
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

func (s *MemoryTransactionStore) load() error {
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
	var items []*model.Transaction
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("corrupt transaction data: %w", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, t := range items {
		if t != nil && t.ID != "" {
			s.items[t.ID] = t
		}
	}
	return nil
}

// MemoryTransferStore is an in-memory transfer store with JSON persistence.
type MemoryTransferStore struct {
	mu    sync.RWMutex
	items []*model.Transfer
	path  string
	dirty bool
	log   *logger.Logger
	stop  chan struct{}
	done  chan struct{}
}

// NewMemoryTransferStore creates a MemoryTransferStore.
func NewMemoryTransferStore(path string, log *logger.Logger, interval time.Duration) (*MemoryTransferStore, error) {
	if log == nil {
		log = logger.Default()
	}
	s := &MemoryTransferStore{
		items: make([]*model.Transfer, 0),
		path:  path,
		log:   log,
		stop:  make(chan struct{}),
		done:  make(chan struct{}),
	}
	if path != "" {
		if err := s.load(); err != nil {
			log.Warn("transfer store load failed", "err", err.Error())
		}
		go s.saveLoop(interval)
	} else {
		close(s.done)
	}
	return s, nil
}

func (s *MemoryTransferStore) Create(ctx context.Context, t *model.Transfer) error {
	if t == nil || t.ID == "" {
		return model.ErrInvalidInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, t.Clone())
	s.dirty = true
	return nil
}

func (s *MemoryTransferStore) ListByAccount(ctx context.Context, accountID string, limit int) ([]*model.Transfer, error) {
	s.mu.RLock()
	var out []*model.Transfer
	for _, t := range s.items {
		if t.FromAccountID == accountID || t.ToAccountID == accountID {
			out = append(out, t.Clone())
		}
	}
	s.mu.RUnlock()
	sort.Slice(out, func(i, j int) bool { return out[i].Date.After(out[j].Date) })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (s *MemoryTransferStore) Close() error {
	if s.path == "" {
		return nil
	}
	s.mu.Lock()
	close(s.stop)
	s.mu.Unlock()
	<-s.done
	return s.flush()
}

func (s *MemoryTransferStore) saveLoop(interval time.Duration) {
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

func (s *MemoryTransferStore) flush() error {
	if s.path == "" {
		return nil
	}
	s.mu.Lock()
	snapshot := make([]*model.Transfer, len(s.items))
	copy(snapshot, s.items)
	s.dirty = false
	s.mu.Unlock()
	sort.Slice(snapshot, func(i, j int) bool { return snapshot[i].Date.After(snapshot[j].Date) })
	payload, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal transfers: %w", err)
	}
	dir := filepath.Dir(s.path)
	if dir != "" && dir != "." {
		os.MkdirAll(dir, 0o755) // #nosec G301
	}
	tmp, err := os.CreateTemp(dir, ".ledgerly-transfers-*.tmp")
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

func (s *MemoryTransferStore) load() error {
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
	var items []*model.Transfer
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("corrupt transfer data: %w", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = items
	return nil
}

// MemoryBudgetStore is an in-memory budget store with JSON persistence.
type MemoryBudgetStore struct {
	mu    sync.RWMutex
	items map[string]*model.Budget
	path  string
	dirty bool
	log   *logger.Logger
	stop  chan struct{}
	done  chan struct{}
}

// NewMemoryBudgetStore creates a MemoryBudgetStore.
func NewMemoryBudgetStore(path string, log *logger.Logger, interval time.Duration) (*MemoryBudgetStore, error) {
	if log == nil {
		log = logger.Default()
	}
	s := &MemoryBudgetStore{
		items: make(map[string]*model.Budget),
		path:  path,
		log:   log,
		stop:  make(chan struct{}),
		done:  make(chan struct{}),
	}
	if path != "" {
		if err := s.load(); err != nil {
			log.Warn("budget store load failed", "err", err.Error())
		}
		go s.saveLoop(interval)
	} else {
		close(s.done)
	}
	return s, nil
}

func (s *MemoryBudgetStore) Create(ctx context.Context, b *model.Budget) error {
	if b == nil || b.ID == "" {
		return model.ErrInvalidInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.items[b.ID]; exists {
		return model.ErrAlreadyExists
	}
	s.items[b.ID] = b.Clone()
	s.dirty = true
	return nil
}

func (s *MemoryBudgetStore) GetByID(ctx context.Context, id string) (*model.Budget, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.items[id]
	if !ok {
		return nil, model.ErrNotFound
	}
	return b.Clone(), nil
}

func (s *MemoryBudgetStore) Update(ctx context.Context, b *model.Budget) error {
	if b == nil || b.ID == "" {
		return model.ErrInvalidInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[b.ID]; !ok {
		return model.ErrNotFound
	}
	s.items[b.ID] = b.Clone()
	s.dirty = true
	return nil
}

func (s *MemoryBudgetStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return model.ErrNotFound
	}
	delete(s.items, id)
	s.dirty = true
	return nil
}

func (s *MemoryBudgetStore) List(ctx context.Context) ([]*model.Budget, error) {
	s.mu.RLock()
	out := make([]*model.Budget, 0, len(s.items))
	for _, b := range s.items {
		out = append(out, b.Clone())
	}
	s.mu.RUnlock()
	sort.Slice(out, func(i, j int) bool { return out[i].CategoryID < out[j].CategoryID })
	return out, nil
}

func (s *MemoryBudgetStore) Close() error {
	if s.path == "" {
		return nil
	}
	s.mu.Lock()
	close(s.stop)
	s.mu.Unlock()
	<-s.done
	return s.flush()
}

func (s *MemoryBudgetStore) saveLoop(interval time.Duration) {
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

func (s *MemoryBudgetStore) flush() error {
	if s.path == "" {
		return nil
	}
	s.mu.Lock()
	snapshot := make([]*model.Budget, 0, len(s.items))
	for _, b := range s.items {
		snapshot = append(snapshot, b.Clone())
	}
	s.dirty = false
	s.mu.Unlock()
	payload, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal budgets: %w", err)
	}
	dir := filepath.Dir(s.path)
	if dir != "" && dir != "." {
		os.MkdirAll(dir, 0o755) // #nosec G301
	}
	tmp, err := os.CreateTemp(dir, ".ledgerly-budgets-*.tmp")
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

func (s *MemoryBudgetStore) load() error {
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
	var items []*model.Budget
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("corrupt budget data: %w", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, b := range items {
		if b != nil && b.ID != "" {
			s.items[b.ID] = b
		}
	}
	return nil
}
