package webhook

import (
	"sync"

	"github.com/google/uuid"
)

type queueLocker struct {
	mu           *sync.Mutex
	whInProgress map[uuid.UUID]struct{}
}

func (q *queueLocker) Acquire(whID uuid.UUID) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	if _, ok := q.whInProgress[whID]; !ok {
		q.whInProgress[whID] = struct{}{}
		return true
	}

	return false
}

func (q *queueLocker) Release(whID uuid.UUID) {
	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.whInProgress, whID)
}
