package rbac

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// Permission of a role to an endpoint
type Permission struct {
	Allow   bool     `yaml:"allow"`
	Ensure  Ensurer  `yaml:"ensure,omitempty"`
	Enforce Enforcer `yaml:"enforce,omitempty"`
}

// Endpoint is a map of {endpoint: permission}
type Endpoint map[string]Permission

// Resource is a map of {resource: endpoint}
type Resource map[string]Endpoint

// RBAC is a map of {role: resource}
type RBAC map[string]Resource

// FromFile creates a new RBAC object from .yaml file
func FromFile(path string) *RBAC {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}

	rbac := &RBAC{}
	err = yaml.Unmarshal(f, rbac)
	if err != nil {
		return nil
	}

	return rbac
}
