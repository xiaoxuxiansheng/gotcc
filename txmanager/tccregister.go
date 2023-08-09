package txmanager

import (
	"github.com/xiaoxuxiansheng/gotcc/component"
)

type TCCRegistryCenter interface {
	Register(component component.TCCComponent) error
	Components(componentIDs ...string) ([]component.TCCComponent, error)
}
