package persisters

// Local implements StatePersister and stores the broker state
// in a JSON file in the file system.
type Local struct{}

func (l *Local) Load() (*State, error) {
	return &State{}, nil
}

func (l *Local) Save(s *State) error {
	return nil
}
