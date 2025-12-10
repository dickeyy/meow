package audio

import (
	"math/rand"
	"sync"
)

type Queue struct {
	tracks   []*Track
	history  []*Track
	position int
	mu       sync.RWMutex
}

func NewQueue() *Queue {
	return &Queue{
		tracks:  make([]*Track, 0),
		history: make([]*Track, 0),
	}
}

func (q *Queue) Add(tracks ...*Track) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tracks = append(q.tracks, tracks...)
}

func (q *Queue) AddNext(track *Track) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.tracks) == 0 {
		q.tracks = append(q.tracks, track)
	} else {
		// Insert at position 1 (after current)
		q.tracks = append(q.tracks[:1], append([]*Track{track}, q.tracks[1:]...)...)
	}
}

func (q *Queue) Current() *Track {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if len(q.tracks) == 0 {
		return nil
	}
	return q.tracks[0]
}

func (q *Queue) Next() *Track {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.tracks) == 0 {
		return nil
	}

	// Move current to history
	current := q.tracks[0]
	q.history = append(q.history, current)

	// Remove from queue
	q.tracks = q.tracks[1:]

	if len(q.tracks) == 0 {
		return nil
	}
	return q.tracks[0]
}

func (q *Queue) Previous() *Track {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.history) == 0 {
		return nil
	}

	// Get last from history
	last := q.history[len(q.history)-1]
	q.history = q.history[:len(q.history)-1]

	// Add back to front of queue
	q.tracks = append([]*Track{last}, q.tracks...)

	return last
}

func (q *Queue) Peek(n int) []*Track {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if n <= 0 || len(q.tracks) <= 1 {
		return nil
	}

	end := n + 1
	if end > len(q.tracks) {
		end = len(q.tracks)
	}

	result := make([]*Track, end-1)
	copy(result, q.tracks[1:end])
	return result
}

func (q *Queue) All() []*Track {
	q.mu.RLock()
	defer q.mu.RUnlock()

	result := make([]*Track, len(q.tracks))
	copy(result, q.tracks)
	return result
}

func (q *Queue) Upcoming() []*Track {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if len(q.tracks) <= 1 {
		return nil
	}

	result := make([]*Track, len(q.tracks)-1)
	copy(result, q.tracks[1:])
	return result
}

func (q *Queue) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.tracks)
}

func (q *Queue) UpcomingLen() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if len(q.tracks) <= 1 {
		return 0
	}
	return len(q.tracks) - 1
}

func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tracks = make([]*Track, 0)
}

func (q *Queue) ClearAll() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tracks = make([]*Track, 0)
	q.history = make([]*Track, 0)
}

func (q *Queue) Remove(index int) *Track {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Index is 1-based for user facing, convert to 0-based
	// But we also skip the currently playing track (index 0)
	actualIndex := index
	if actualIndex < 1 || actualIndex >= len(q.tracks) {
		return nil
	}

	removed := q.tracks[actualIndex]
	q.tracks = append(q.tracks[:actualIndex], q.tracks[actualIndex+1:]...)
	return removed
}

func (q *Queue) Move(from, to int) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Indices are 1-based for users, but we skip current (index 0)
	if from < 1 || from >= len(q.tracks) || to < 1 || to >= len(q.tracks) {
		return false
	}

	if from == to {
		return true
	}

	track := q.tracks[from]
	q.tracks = append(q.tracks[:from], q.tracks[from+1:]...)

	if to > from {
		to--
	}
	q.tracks = append(q.tracks[:to], append([]*Track{track}, q.tracks[to:]...)...)

	return true
}

func (q *Queue) Shuffle() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.tracks) <= 2 {
		return
	}

	// Keep current track, shuffle the rest
	upcoming := q.tracks[1:]
	rand.Shuffle(len(upcoming), func(i, j int) {
		upcoming[i], upcoming[j] = upcoming[j], upcoming[i]
	})
}

func (q *Queue) IsEmpty() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.tracks) == 0
}

func (q *Queue) HasNext() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.tracks) > 1
}

func (q *Queue) HasPrevious() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.history) > 0
}

