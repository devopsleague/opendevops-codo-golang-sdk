package dummy

import (
	"context"
	"github.com/go-kratos/kratos/v2/registry"
)

var _ registry.Registrar = (*Registrar)(nil)

type Registrar struct {
}

func NewRegistrar() *Registrar {
	return &Registrar{}
}

func (x *Registrar) Register(ctx context.Context, service *registry.ServiceInstance) error {
	return nil
}

func (x *Registrar) Deregister(ctx context.Context, service *registry.ServiceInstance) error {
	return nil
}
