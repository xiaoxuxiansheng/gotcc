package example

import (
	"context"

	"github.com/xiaoxuxiansheng/gotcc/component"
)

type MockRegister struct {
}

func (m *MockRegister) Register(ctx context.Context, component component.TCCComponent) error {
	return nil
}

func (m *MockRegister) Components(ctx context.Context, componentIDs ...string) ([]component.TCCComponent, error) {
	return nil, nil
}
