package codegen

import (
	"fmt"
	"sync"
)

var (
	generators = make(map[string]CodeGenerator)
	mu         sync.RWMutex
)

// Register registers a code generator
func Register(generator CodeGenerator) {
	mu.Lock()
	defer mu.Unlock()
	generators[generator.Name()] = generator
}

// Get returns a code generator by name
func Get(name string) (CodeGenerator, error) {
	mu.RLock()
	defer mu.RUnlock()

	generator, ok := generators[name]
	if !ok {
		return nil, fmt.Errorf("generator not found: %s", name)
	}

	return generator, nil
}

// List returns all registered generator names
func List() []string {
	mu.RLock()
	defer mu.RUnlock()

	names := make([]string, 0, len(generators))
	for name := range generators {
		names = append(names, name)
	}
	return names
}
