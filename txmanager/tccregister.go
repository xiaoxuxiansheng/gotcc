package txmanager

import (
	"errors"
	"fmt"
	"sync"

	"github.com/xiaoxuxiansheng/gotcc/component"
)

type registryCenter struct {
	mux        sync.RWMutex
	components map[string]component.TCCComponent
}

func newRegistryCenter() *registryCenter {
	return &registryCenter{
		components: make(map[string]component.TCCComponent),
	}
}

func (r *registryCenter) register(component component.TCCComponent) error {
	r.mux.Lock()
	defer r.mux.Unlock()
	if _, ok := r.components[component.ID()]; ok {
		return errors.New("repeat component id")
	}
	r.components[component.ID()] = component
	return nil
}

func (r *registryCenter) getComponents(componentIDs ...string) ([]component.TCCComponent, error) {
	components := make([]component.TCCComponent, 0, len(componentIDs))

	r.mux.RLock()
	defer r.mux.RUnlock()

	for _, componentID := range componentIDs {
		component, ok := r.components[componentID]
		if !ok {
			return nil, fmt.Errorf("component id: %s not existed", componentID)
		}
		components = append(components, component)
	}

	return components, nil
}
