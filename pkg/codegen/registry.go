package codegen

import "fmt"

var (
	generators = make(map[string]CodeGenerator)
)

// Register registers a code generator
func Register(generator CodeGenerator) {
	generators[generator.Name()] = generator
}

// Get returns a code generator by name
func Get(name string) (CodeGenerator, error) {
	generator, ok := generators[name]
	if !ok {
		return nil, fmt.Errorf("generator not found: %s", name)
	}

	return generator, nil
}

// List returns all registered generator names
func List() []string {
	names := make([]string, 0, len(generators))
	for name := range generators {
		names = append(names, name)
	}
	return names
}
