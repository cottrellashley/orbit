package port

import "github.com/cottrellashley/orbit/internal/domain"

// ProjectRepositoryFromEnvRepo returns a ProjectRepository backed by a
// legacy EnvironmentRepository. It converts between domain.Environment
// and domain.Project on every call, providing a transparent migration
// bridge so new code can program against ProjectRepository while the
// underlying storage still speaks Environment.
//
// Topology and integration data are not preserved through the round-trip
// because EnvironmentRepository has no fields for them. Callers needing
// those fields should use a native ProjectRepository adapter.
func ProjectRepositoryFromEnvRepo(envRepo EnvironmentRepository) ProjectRepository {
	return &envProjectBridge{envRepo: envRepo}
}

type envProjectBridge struct {
	envRepo EnvironmentRepository
}

func (b *envProjectBridge) List() ([]*domain.Project, error) {
	envs, err := b.envRepo.List()
	if err != nil {
		return nil, err
	}
	projects := make([]*domain.Project, len(envs))
	for i, env := range envs {
		projects[i] = domain.ProjectFromEnvironment(env)
	}
	return projects, nil
}

func (b *envProjectBridge) Get(name string) (*domain.Project, error) {
	env, err := b.envRepo.Get(name)
	if err != nil {
		return nil, err
	}
	return domain.ProjectFromEnvironment(env), nil
}

func (b *envProjectBridge) GetByPath(path string) (*domain.Project, error) {
	env, err := b.envRepo.GetByPath(path)
	if err != nil {
		return nil, err
	}
	return domain.ProjectFromEnvironment(env), nil
}

func (b *envProjectBridge) Save(projects []*domain.Project) error {
	envs := make([]*domain.Environment, len(projects))
	for i, p := range projects {
		envs[i] = domain.EnvironmentFromProject(p)
	}
	return b.envRepo.Save(envs)
}

func (b *envProjectBridge) Delete(name string) error {
	return b.envRepo.Delete(name)
}
