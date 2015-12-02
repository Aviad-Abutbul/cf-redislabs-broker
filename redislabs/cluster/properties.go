package cluster

// InstanceSettings is a JSON serializable collection of properties necessary
// for the creation of a cluster instance (database).
type InstanceSettings struct {
	MemoryLimit int64 `json:"memory_size"`
	Replication bool  `json:"replication"`
	ShardCount  int64 `json:"shards_count"`
}

// InstanceCredentials contains properties necessary for identifying a
// cluster instance (database) and connecting to it.
type InstanceCredentials struct {
	UID      int
	Port     int
	IPList   []string
	Password string
}
