package persisters

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
)

var (
	BrokerFolder  = ".redislabs-broker"
	StateFile     = "state.json"
	StateFileMask = os.FileMode(0777)
)

// Local implements StatePersister and stores the broker state
// in a JSON file in the file system.
type Local struct{}

// Load loads the state from the local JSON file. If such file
// does not exist (no Save has been made to date) it returns an empty state.
func (l *Local) Load() (*State, error) {
	p, err := GetStatePath()
	// Return an empty state if the file does not exist.
	if _, err = os.Stat(p); os.IsNotExist(err) {
		return &State{}, nil
	}
	bytes, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}
	s := State{}
	err = json.Unmarshal(bytes, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Save saves the state to the local JSON file. It creates it if it does not
// exist.
func (l *Local) Save(s *State) error {
	p, err := GetStatePath()
	if err != nil {
		return err
	}
	bytes, err := json.Marshal(s)
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(p, bytes, StateFileMask); err != nil {
		return err
	}
	return nil
}

// GetStatePath returns the file path for the state file. The function
// should be overriden in tests.
var GetStatePath = func() (string, error) {
	home := path.Join(os.Getenv("HOME"), BrokerFolder)
	// Create home directory if it does not exist.
	if _, err := os.Stat(home); os.IsNotExist(err) {
		if err := os.MkdirAll(home, StateFileMask); err != nil {
			return "", err
		}
	}
	return path.Join(home, StateFile), nil
}
