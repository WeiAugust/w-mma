package review

import (
	"context"
	"errors"
	"sort"
	"sync"
)

// MemoryRepository is an in-memory implementation used for local MVP linking.
type MemoryRepository struct {
	mu sync.Mutex

	nextPendingID int64
	pending       map[int64]PendingArticle
	published     []PendingArticle
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		nextPendingID: 1,
		pending: map[int64]PendingArticle{
			1: {
				ID:        1,
				SourceID:  1,
				Title:     "UFC 300 主赛前瞻",
				Summary:   "主赛阵容与战术看点速览",
				SourceURL: "https://www.ufc.com",
			},
		},
		published: []PendingArticle{},
	}
}

func (m *MemoryRepository) GetPending(_ context.Context, pendingID int64) (PendingArticle, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, ok := m.pending[pendingID]
	if !ok {
		return PendingArticle{}, errors.New("pending item not found")
	}
	return item, nil
}

func (m *MemoryRepository) PublishArticle(_ context.Context, rec PendingArticle) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.published = append(m.published, rec)
	return nil
}

func (m *MemoryRepository) MarkApproved(_ context.Context, pendingID int64, _ int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.pending, pendingID)
	return nil
}

func (m *MemoryRepository) ListPending(context.Context) ([]PendingArticle, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	items := make([]PendingArticle, 0, len(m.pending))
	for _, item := range m.pending {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items, nil
}

func (m *MemoryRepository) CreatePending(_ context.Context, item PendingArticle) (PendingArticle, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	item.ID = m.nextPendingID
	m.nextPendingID++
	m.pending[item.ID] = item
	return item, nil
}

func (m *MemoryRepository) ListPublished(context.Context) ([]PendingArticle, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	items := make([]PendingArticle, len(m.published))
	copy(items, m.published)
	return items, nil
}
