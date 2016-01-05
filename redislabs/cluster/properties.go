package cluster

// InstanceCredentials contains properties necessary for identifying a
// cluster instance (database) and connecting to it.
type InstanceCredentials struct {
	UID      int
	Port     int
	IPList   []string
	Password string
}
