// This file provides a fan-out store used to keep the file-backed control
// plane as the primary local runtime while mirroring state into optional
// secondary stores such as PostgreSQL.
package orchestration

// MultiStore writes run state to multiple underlying stores.
type MultiStore struct {
	primary Store
	mirrors []Store
}

// NewMultiStore constructs a primary-plus-mirrors store.
func NewMultiStore(primary Store, mirrors ...Store) Store {
	filtered := make([]Store, 0, len(mirrors))
	for _, mirror := range mirrors {
		if mirror != nil {
			filtered = append(filtered, mirror)
		}
	}
	return &MultiStore{
		primary: primary,
		mirrors: filtered,
	}
}

// ListPipelineRuns reads from the primary store.
func (s *MultiStore) ListPipelineRuns() ([]PipelineRun, error) {
	return s.primary.ListPipelineRuns()
}

// SavePipelineRun writes to the primary store first, then mirrors.
func (s *MultiStore) SavePipelineRun(run PipelineRun) error {
	if err := s.primary.SavePipelineRun(run); err != nil {
		return err
	}
	for _, mirror := range s.mirrors {
		_ = mirror.SavePipelineRun(run)
	}
	return nil
}

// GetPipelineRun reads from the primary store.
func (s *MultiStore) GetPipelineRun(id string) (PipelineRun, bool, error) {
	return s.primary.GetPipelineRun(id)
}
