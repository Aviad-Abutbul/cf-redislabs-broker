package cluster

// InstanceSettings is a collection of properties needed to create
// a cluster instance.
type InstanceSettings struct {
	MemoryLimit int64
	Replication bool
	ShardCount  int64
}
