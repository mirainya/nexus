package pipeline

import (
	"fmt"
	"sync"
)

var (
	registry = make(map[string]Processor)
	mu       sync.RWMutex
)

func Register(p Processor) {
	mu.Lock()
	defer mu.Unlock()
	registry[p.Name()] = p
}

func Get(name string) (Processor, error) {
	mu.RLock()
	defer mu.RUnlock()
	p, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("processor not found: %s", name)
	}
	return p, nil
}

func All() map[string]Processor {
	mu.RLock()
	defer mu.RUnlock()
	out := make(map[string]Processor, len(registry))
	for k, v := range registry {
		out[k] = v
	}
	return out
}
