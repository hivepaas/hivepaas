package settingmountserviceimpl

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/docker"
)

// fakeDocker keeps secrets, configs and one service in memory, and filters by
// label as the daemon does: "key" is present, "key=value" matches.
type fakeDocker struct {
	docker.Manager
	secrets map[string]swarm.Secret
	configs map[string]swarm.Config
	service *swarm.Service
	nextID  int
	created []string
	updates int
	// inUse is how many more removals of an id fail, as a secret a service still
	// references does.
	inUse map[string]int
	// gone makes the service answer as one that does not exist.
	gone bool
}

func newFakeDocker(spec swarm.ServiceSpec) *fakeDocker {
	if spec.TaskTemplate.ContainerSpec == nil {
		spec.TaskTemplate.ContainerSpec = &swarm.ContainerSpec{}
	}
	return &fakeDocker{secrets: map[string]swarm.Secret{}, configs: map[string]swarm.Config{},
		service: &swarm.Service{ID: "svc_1", Spec: spec}, inUse: map[string]int{}}
}

func matches(labels map[string]string, filters client.Filters) bool {
	for want := range filters["label"] {
		key, value, withValue := strings.Cut(want, "=")
		got, ok := labels[key]
		if !ok || (withValue && got != value) {
			return false
		}
	}
	return true
}

func (f *fakeDocker) id(kind string) string {
	f.nextID++
	return fmt.Sprintf("%s_%d", kind, f.nextID)
}

func (f *fakeDocker) SecretList(
	_ context.Context, options ...docker.SecretListOption,
) (*client.SecretListResult, error) {
	opts := client.SecretListOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	out := &client.SecretListResult{}
	for _, secret := range f.secrets {
		if matches(secret.Spec.Labels, opts.Filters) {
			out.Items = append(out.Items, secret)
		}
	}
	return out, nil
}

func (f *fakeDocker) SecretCreate(
	_ context.Context, name string, data []byte, options ...docker.SecretCreateOption,
) (*client.SecretCreateResult, error) {
	for _, secret := range f.secrets {
		if secret.Spec.Name == name {
			return nil, hperrors.Wrap(hperrors.ErrInfraAlreadyExists)
		}
	}
	opts := client.SecretCreateOptions{}
	opts.Spec.Name, opts.Spec.Data = name, data
	for _, opt := range options {
		opt(&opts)
	}
	id := f.id("secret")
	f.secrets[id] = swarm.Secret{ID: id, Spec: opts.Spec}
	f.created = append(f.created, name)
	return &client.SecretCreateResult{ID: id}, nil
}

func (f *fakeDocker) SecretRemove(
	_ context.Context, id string, _ ...docker.SecretRemoveOption,
) (*client.SecretRemoveResult, error) {
	if f.inUse[id] > 0 {
		f.inUse[id]--
		return nil, hperrors.Wrap(hperrors.ErrInfraConflict)
	}
	delete(f.secrets, id)
	return &client.SecretRemoveResult{}, nil
}

func (f *fakeDocker) ConfigList(
	_ context.Context, options ...docker.ConfigListOption,
) (*client.ConfigListResult, error) {
	opts := client.ConfigListOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	out := &client.ConfigListResult{}
	for _, config := range f.configs {
		if matches(config.Spec.Labels, opts.Filters) {
			out.Items = append(out.Items, config)
		}
	}
	return out, nil
}

func (f *fakeDocker) ConfigCreate(
	_ context.Context, name string, data []byte, options ...docker.ConfigCreateOption,
) (*client.ConfigCreateResult, error) {
	for _, config := range f.configs {
		if config.Spec.Name == name {
			return nil, hperrors.Wrap(hperrors.ErrInfraAlreadyExists)
		}
	}
	opts := client.ConfigCreateOptions{}
	opts.Spec.Name, opts.Spec.Data = name, data
	for _, opt := range options {
		opt(&opts)
	}
	id := f.id("config")
	f.configs[id] = swarm.Config{ID: id, Spec: opts.Spec}
	f.created = append(f.created, name)
	return &client.ConfigCreateResult{ID: id}, nil
}

func (f *fakeDocker) ConfigRemove(
	_ context.Context, id string, _ ...docker.ConfigRemoveOption,
) (*client.ConfigRemoveResult, error) {
	if f.inUse[id] > 0 {
		f.inUse[id]--
		return nil, hperrors.Wrap(hperrors.ErrInfraConflict)
	}
	delete(f.configs, id)
	return &client.ConfigRemoveResult{}, nil
}

func (f *fakeDocker) ServiceInspect(
	_ context.Context, _ string, _ ...docker.ServiceInspectOption,
) (*client.ServiceInspectResult, error) {
	if f.gone {
		return nil, hperrors.NewNotFound("Service")
	}
	return &client.ServiceInspectResult{Service: *f.copyService()}, nil
}

func (f *fakeDocker) ServiceUpdateFunc(
	_ context.Context, _ string, _ *swarm.Service, fn func(int, *swarm.Service) (bool, error),
	_ int, _ time.Duration, _ ...docker.ServiceUpdateOption,
) error {
	svc := f.copyService()
	changed, err := fn(0, svc)
	if err != nil {
		return err
	}
	if changed {
		f.service = svc
		f.updates++
	}
	return nil
}

// copyService is the service as a fresh inspect returns it.
func (f *fakeDocker) copyService() *swarm.Service {
	svc := *f.service
	contSpec := *svc.Spec.TaskTemplate.ContainerSpec
	contSpec.Secrets = append([]*swarm.SecretReference(nil), contSpec.Secrets...)
	contSpec.Configs = append([]*swarm.ConfigReference(nil), contSpec.Configs...)
	svc.Spec.TaskTemplate.ContainerSpec = &contSpec
	return &svc
}
