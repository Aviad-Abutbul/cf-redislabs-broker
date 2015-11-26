package persisters

// StatePersister is responsible for saving & retrieving
// the broker state, the information about available service
// instances and their parameters.
type StatePersister interface {
	Save(s *State) error
	Load() (*State, error)
}

type State struct {
	AvailableInstances []ServiceInstance
}

type ServiceInstance struct {
	ID       string
	Port     int64
	IPList   []string
	Password string
}
