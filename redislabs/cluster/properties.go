package cluster

import (
	"encoding/json"
	"reflect"
	"strings"
)

// InstanceSettings is a JSON serializable collection of properties necessary
// for the creation of a cluster instance (database).
type InstanceSettings struct {
	MemoryLimit int64  `json:"memory_size"`
	Replication bool   `json:"replication"`
	ShardCount  int64  `json:"shards_count"`
	Password    string `json:"authentication_redis_pass"`
}

// InstanceCredentials contains properties necessary for identifying a
// cluster instance (database) and connecting to it.
type InstanceCredentials struct {
	UID      int
	Port     int
	IPList   []string
	Password string
}

// updateParameters serves as a contract for additional cluser properties
// allowed to be updated.
type updateParameters struct {
	MemoryLimit int64 `json:"memory_size"`
}

// CheckUpdateParameters verifies the contract for additional (not coming from
// a plan change) properties of the cluster allowed to be updated.
func CheckUpdateParameters(params map[string]interface{}) error {
	updatables := updateParametersJSONKeys()
	for k, v := range params {
		if !contains(updatables, k) {
			return ErrUnknownParam(k)
		}
		bytes, err := json.Marshal(map[string]interface{}{k: v})
		if err != nil {
			return ErrInvalidJSON
		}
		if err = json.Unmarshal(bytes, &updateParameters{}); err != nil {
			return ErrInvalidType(k)
		}
	}
	return nil
}

func updateParametersJSONKeys() []string {
	keys := []string{}
	t := reflect.TypeOf(updateParameters{})
	for i := 0; i < t.NumField(); i++ {
		f := t.FieldByIndex([]int{i})
		keys = append(keys, strings.Split(f.Tag.Get("json"), ",")[0])
	}
	return keys
}

func contains(strs []string, str string) bool {
	for _, s := range strs {
		if s == str {
			return true
		}
	}
	return false
}
