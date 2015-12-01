package persisters

import "github.com/Altoros/cf-redislabs-broker/redislabs/cluster"

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
	ID          string
	Credentials cluster.InstanceCredentials
}
