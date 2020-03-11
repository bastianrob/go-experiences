package mock

// CRUD contract
type CRUD interface {
	Get(id string) (interface{}, error)
	Create(dao interface{}) error
	Update(dao interface{}) error
}

// APIClient generic mock implementation of CRUD interface
type APIClient struct {
	GetFunc    func(id string) (interface{}, error)
	CreateFunc func(dao interface{}) error
	UpdateFunc func(dao interface{}) error
}

// Get mock, please implement GetFunc
func (ac *APIClient) Get(id string) (interface{}, error) {
	return ac.GetFunc(id)
}

// Create mock, please implement CreateFunc
func (ac *APIClient) Create(dao interface{}) error {
	return ac.CreateFunc(dao)
}

// Update mock, please implement UpdateFunc
func (ac *APIClient) Update(dao interface{}) error {
	return ac.UpdateFunc(dao)
}
