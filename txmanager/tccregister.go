package txmanager

import (
	"context"

	"github.com/xiaoxuxiansheng/gotcc/component"
)

type TCCRegistyCenter interface {
	Register(ctx context.Context, component component.TCCComponent) error
	Components(ctx context.Context, componentIDs ...string) ([]component.TCCComponent, error)
}
