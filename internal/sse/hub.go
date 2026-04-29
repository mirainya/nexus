package sse

import (
	"encoding/json"
	"sync"
)

type Event struct {
	Type      string `json:"type"`
	Step      int    `json:"step,omitempty"`
	Processor string `json:"processor,omitempty"`
	Data      any    `json:"data,omitempty"`
	Error     string `json:"error,omitempty"`
}

type Hub struct {
	mu   sync.RWMutex
	subs map[string][]chan Event
}

var defaultHub = &Hub{subs: make(map[string][]chan Event)}

func Default() *Hub { return defaultHub }

func (h *Hub) Subscribe(jobUUID string) chan Event {
	ch := make(chan Event, 32)
	h.mu.Lock()
	h.subs[jobUUID] = append(h.subs[jobUUID], ch)
	h.mu.Unlock()
	return ch
}

func (h *Hub) Unsubscribe(jobUUID string, ch chan Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	channels := h.subs[jobUUID]
	for i, c := range channels {
		if c == ch {
			h.subs[jobUUID] = append(channels[:i], channels[i+1:]...)
			break
		}
	}
	if len(h.subs[jobUUID]) == 0 {
		delete(h.subs, jobUUID)
	}
	close(ch)
}

func (h *Hub) Publish(jobUUID string, evt Event) {
	h.mu.RLock()
	channels := h.subs[jobUUID]
	h.mu.RUnlock()
	for _, ch := range channels {
		select {
		case ch <- evt:
		default:
		}
	}
}

func (e Event) JSON() []byte {
	b, _ := json.Marshal(e)
	return b
}
