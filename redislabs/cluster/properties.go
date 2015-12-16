package cluster

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
)

// InstanceSettings is a JSON serializable collection of properties necessary
// for the creation of a cluster instance (database).
type InstanceSettings struct {
	Name     string `json:"name"`
	Password string `json:"authentication_redis_pass"`
	PlanSettings
}

// PlanSettings is a JSON serializable collection of properties that have to
// be provided by a service plan.
type PlanSettings struct {
	MemoryLimit      int64               `json:"memory_size"`
	Replication      bool                `json:"replication"`
	ShardCount       int64               `json:"shards_count"`
	Sharding         bool                `json:"sharding"`
	ImplicitShardKey bool                `json:"implicit_shard_key"`
	ShardKeyRegex    []map[string]string `json:"shard_key_regex,omitempty"`
	Persistence      string              `json:"data_persistence,omitempty"`
	Snapshot         []Snapshot          `json:"snapshot_policy,omitempty"`
}

type Snapshot struct {
	Writes int `json:"writes"`
	Secs   int `json:"secs"`
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
	MemoryLimit int64      `json:"memory_size"`
	Persistence string     `json:"data_persistence"`
	Snapshot    []Snapshot `json:"snapshot_policy"`
}

// CheckUpdateParameters verifies the contract for additional (not coming from
// a plan change) properties of the cluster allowed to be updated. Also,
// CheckUpdateParameters returns a map with some values representation fixes.
func CheckUpdateParameters(params map[string]interface{}) (map[string]interface{}, error) {
	updatables := updateParametersJSONKeys()
	for k, v := range params {
		if !contains(updatables, k) {
			return params, ErrUnknownParam(k)
		}
		if k == "memory_size" {
			// This is a dirty hack for the purpose of transforming
			// the number parsed in the exponent format back to the non-exponent format.
			// We do not want to make brokerapi parse numbers into json.Numbers to
			// avoid extra deviations from master (the need for json.Numbers is unlikely
			// to be the common case). Neither do we want to support reflection-based
			// mess for checking types.
			fv, isFloat := v.(float64)
			if isFloat {
				s := strconv.FormatFloat(fv, 'f', -1, 64)
				i, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					return params, ErrInvalidType(k)
				}
				v = i
				params[k] = v
			}
		}
		bytes, err := json.Marshal(map[string]interface{}{k: v})
		if err != nil {
			return params, ErrInvalidJSON
		}
		if err = json.Unmarshal(bytes, &updateParameters{}); err != nil {
			return params, ErrInvalidType(k)
		}
	}
	return params, nil
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
