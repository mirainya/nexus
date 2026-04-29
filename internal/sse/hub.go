package sse

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/mirainya/nexus/pkg/config"
	"github.com/mirainya/nexus/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
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
	rdb  *redis.Client
}

var defaultHub *Hub

func Init() {
	h := &Hub{subs: make(map[string][]chan Event)}

	cfg := config.C.Redis
	if cfg.Addr != "" {
		h.rdb = redis.NewClient(&redis.Options{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.DB,
		})
		go h.subscribeRedis()
	}

	defaultHub = h
}

func Default() *Hub {
	if defaultHub == nil {
		defaultHub = &Hub{subs: make(map[string][]chan Event)}
	}
	return defaultHub
}

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
	if h.rdb != nil {
		msg := struct {
			JobUUID string `json:"job_uuid"`
			Event   Event  `json:"event"`
		}{JobUUID: jobUUID, Event: evt}
		if b, err := json.Marshal(msg); err == nil {
			h.rdb.Publish(context.Background(), "nexus:sse", string(b)).Err()
		}
		return
	}
	h.localPublish(jobUUID, evt)
}

func (h *Hub) localPublish(jobUUID string, evt Event) {
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

func (h *Hub) subscribeRedis() {
	pubsub := h.rdb.Subscribe(context.Background(), "nexus:sse")
	defer pubsub.Close()

	for msg := range pubsub.Channel() {
		var parsed struct {
			JobUUID string `json:"job_uuid"`
			Event   Event  `json:"event"`
		}
		if err := json.Unmarshal([]byte(msg.Payload), &parsed); err != nil {
			logger.Warn("failed to parse sse redis message", zap.Error(err))
			continue
		}
		h.localPublish(parsed.JobUUID, parsed.Event)
	}
}

func (e Event) JSON() []byte {
	b, _ := json.Marshal(e)
	return b
}
