package genetics

import (
	"sync"
	"sync/atomic"
)

// innovationKey identifies a unique structural change for deduplication.
// For link innovations oldInnovNum is 0; for node innovations isRecurrent is always false.
type innovationKey struct {
	inNodeId, outNodeId int
	typ                 innovationType
	isRecurrent         bool
	oldInnovNum         int64
}

// InnovationTracker is the Population's implementation of InnovationsObserver and
// network.NodeIdGenerator. It uses a map for O(1) innovation lookup so that a
// long-running ALife simulation never suffers O(n) linear scans.
type InnovationTracker struct {
	mu           sync.RWMutex
	byKey        map[innovationKey]*Innovation
	nextInnovNum int64
	nextNodeId   int32
}

func newInnovationTracker() *InnovationTracker {
	return &InnovationTracker{
		byKey: make(map[innovationKey]*Innovation),
	}
}

// --- InnovationsObserver ---

func (t *InnovationTracker) StoreInnovation(inn Innovation) {
	key := keyFor(&inn)
	t.mu.Lock()
	t.byKey[key] = &inn
	t.mu.Unlock()
}

func (t *InnovationTracker) Innovations() []Innovation {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]Innovation, 0, len(t.byKey))
	for _, v := range t.byKey {
		out = append(out, *v)
	}
	return out
}

func (t *InnovationTracker) NextInnovationNumber() int64 {
	return atomic.AddInt64(&t.nextInnovNum, 1)
}

func (t *InnovationTracker) FindLinkInnovation(inNodeId, outNodeId int, isRecurrent bool) *Innovation {
	key := innovationKey{
		inNodeId:    inNodeId,
		outNodeId:   outNodeId,
		typ:         newLinkInnType,
		isRecurrent: isRecurrent,
	}
	t.mu.RLock()
	inn := t.byKey[key]
	t.mu.RUnlock()
	return inn
}

func (t *InnovationTracker) FindNodeInnovation(inNodeId, outNodeId int, oldInnovNum int64) *Innovation {
	key := innovationKey{
		inNodeId:    inNodeId,
		outNodeId:   outNodeId,
		typ:         newNodeInnType,
		oldInnovNum: oldInnovNum,
	}
	t.mu.RLock()
	inn := t.byKey[key]
	t.mu.RUnlock()
	return inn
}

// --- network.NodeIdGenerator ---

func (t *InnovationTracker) NextNodeId() int {
	return int(atomic.AddInt32(&t.nextNodeId, 1))
}

// clear resets all stored innovations (does not reset counters).
func (t *InnovationTracker) clear() {
	t.mu.Lock()
	t.byKey = make(map[innovationKey]*Innovation)
	t.mu.Unlock()
}

// keyFor derives the lookup key for an innovation.
func keyFor(inn *Innovation) innovationKey {
	if inn.innovationType == newNodeInnType {
		return innovationKey{
			inNodeId:    inn.InNodeId,
			outNodeId:   inn.OutNodeId,
			typ:         newNodeInnType,
			oldInnovNum: inn.OldInnovNum,
		}
	}
	return innovationKey{
		inNodeId:    inn.InNodeId,
		outNodeId:   inn.OutNodeId,
		typ:         newLinkInnType,
		isRecurrent: inn.IsRecurrent,
	}
}
