package adapters

type Default struct{}

func (d *Default) Create(instanceID string) error {
	return nil
}

func (d *Default) Destroy(instanceID string) error {
	return nil
}

func (d *Default) InstanceExists(instanceID string) (bool, error) {
	return false, nil
}
