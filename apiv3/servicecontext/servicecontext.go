package servicecontext

import (
	"github.com/evergreen-ci/evergreen"
)

// ServiceContext is a struct that contains all of the methods which
// connect to the service layer of evergreen. These methods abstract the link
// between the service and the API layers, allowing for changes in the
// service architecture without forcing changes to the API.
type ServiceContext struct {
	Settings evergreen.Settings

	TaskConnector
	UserConnector
	ContextConnector
}

// NewServiceContext returns a ServiceContext with interface implementations
// that connect directly to the underlying service layer.
func NewServiceContext() ServiceContext {
	return ServiceContext{
		TaskConnector:    &DBTaskConnector{},
		UserConnector:    &DBUserConnector{},
		ContextConnector: &DBContextConnector{},
	}
}

// NewServiceContext returns a ServiceContext with interface implementations
// that mock connecting to the service layer.
func NewMockServiceContext() ServiceContext {
	return ServiceContext{
		TaskConnector:    &MockTaskConnector{},
		UserConnector:    &MockUserConnector{},
		ContextConnector: &MockContextConnector{},
	}

}
