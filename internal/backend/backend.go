package backend

import (
	"context"
	"fmt"

	"github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/vkms"
)

type Instance struct {
	Name string
	Path string
}

type Backend interface {
	Name() string
	Create(context.Context, string) (Instance, error)
	Destroy(string) error
}

type VKMSBackend struct{ Manager *vkms.Manager }

func NewVKMSBackend(manager *vkms.Manager) *VKMSBackend { return &VKMSBackend{Manager: manager} }
func (b *VKMSBackend) Name() string                     { return "vkms" }
func (b *VKMSBackend) Create(ctx context.Context, name string) (Instance, error) {
	inst, err := b.Manager.Create(ctx, name)
	if err != nil {
		return Instance{}, err
	}
	return Instance{Name: inst.Name, Path: inst.Path}, nil
}
func (b *VKMSBackend) Destroy(name string) error { return b.Manager.Destroy(name) }

type ExperimentalPortalBackend struct{}

func NewExperimentalPortalBackend() *ExperimentalPortalBackend { return &ExperimentalPortalBackend{} }
func (b *ExperimentalPortalBackend) Name() string              { return "portal" }
func (b *ExperimentalPortalBackend) Create(context.Context, string) (Instance, error) {
	return Instance{}, fmt.Errorf("portal backend is not implemented")
}
func (b *ExperimentalPortalBackend) Destroy(string) error { return nil }
