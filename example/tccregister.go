package example

import (
	"errors"
	"fmt"
	"sync"

	"github.com/xiaoxuxiansheng/gotcc/component"
)

type MockRegistryCenter struct {
	mux        sync.RWMutex
	components map[string]component.TCCComponent
}

func NewMockRegistryCenter() *MockRegistryCenter {
	return &MockRegistryCenter{
		components: make(map[string]component.TCCComponent),
	}
}

func (m *MockRegistryCenter) Register(component component.TCCComponent) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	if _, ok := m.components[component.ID()]; ok {
		return errors.New("repeat component id")
	}
	m.components[component.ID()] = component
	return nil
}

func (m *MockRegistryCenter) Components(componentIDs ...string) ([]component.TCCComponent, error) {
	components := make([]component.TCCComponent, 0, len(componentIDs))

	m.mux.RLock()
	defer m.mux.RUnlock()

	for _, componentID := range componentIDs {
		component, ok := m.components[componentID]
		if !ok {
			return nil, fmt.Errorf("component id: %s not existed", componentID)
		}
		components = append(components, component)
	}

	return components, nil
}
