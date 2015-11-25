package adapters

type DefaultCreator struct {}

func (d *DefaultCreator) Create(instanceID string) error {
    return nil
}

func (d *DefaultCreator) Destroy(instanceID string) error {
	return nil
}

func (d *DefaultCreator) InstanceExists(instanceID string) (bool, error) {
	return false, nil
}
