package persisters

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

var (
	stateFileMask = os.FileMode(0777)
)

// Local implements StatePersister and stores the broker state
// in a JSON file in the file system.
type local struct {
	stateFilePath string
	lock          sync.Mutex
}

// Load loads the state from the local JSON file. If such file
// does not exist (no Save has been made to date) it returns an empty state.
func (l *local) Load() (*State, error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	// Return an empty state if the file does not exist.
	if _, err := os.Stat(l.stateFilePath); os.IsNotExist(err) {
		return &State{}, nil
	}

	bytes, err := ioutil.ReadFile(l.stateFilePath)
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
func (l *local) Save(s *State) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	stateFileFolder, err := filepath.Abs(filepath.Dir(l.stateFilePath))
	if err != nil {
		return err
	}

	err = os.MkdirAll(stateFileFolder, stateFileMask)
	if err != nil {
		return err
	}

	bytes, err := json.Marshal(s)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(l.stateFilePath, bytes, stateFileMask)
	if err != nil {
		return err
	}
	return nil
}

func NewLocalPersister(path string) StatePersister {
	return &local{
		stateFilePath: path,
	}
}
