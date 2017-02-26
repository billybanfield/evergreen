package servicecontext

import (
	"github.com/evergreen-ci/evergreen/model"
)

// ContextConnector is an interface which abstracts the service layer's fetching
// of a project context.
type ContextConnector interface {

	// FetchContext is a method to fetch a context given a series of identifiers.
	FetchContext(string, string, string, string, string) (model.Context, error)
}

// DBContextConnector is a struct that implements the ContextConnector interface
// through interactions with the backing database.
type DBContextConnector struct{}

// LoadContext fetches the context through a call to the service layer.
func (dc *DBContextConnector) FetchContext(taskId, buildId, versionId, patchId, projectId string) (model.Context, error) {
	return model.LoadContext(taskId, buildId, versionId, patchId, projectId)
}

// MockContextConnector is a struct that implements the ContextConnector interface
// by storing context to be fetched by its method.
type MockContextConnector struct {
	CachedContext model.Context
	CachedErr     error
}

// FetchContext returns the context cached within the MockContextConnector.
func (mc *MockContextConnector) FetchContext(taskId, buildId, versionId, patchId, projectId string) (model.Context, error) {
	return mc.CachedContext, mc.CachedErr
}
