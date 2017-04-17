package route

import (
	"github.com/evergreen/evergreen-ci/apiv3"
	"github.com/evergreen/evergreen-ci/apiv3/servicecontext"
)

func getTestRouteManager(route string, version int) *RouteManager {
	tep := &testGetHandler{}
	testGetMethodHandler := MethodHandler{
		PrefetchFunctions: []PrefetchFunc{PrefetchUser, PrefetchProjectContext},
		Authenticator:     &RequireUserAuthenticator{},
		RequestHandler:    tep.Handler(),
		MethodType:        evergreen.MethodGet,
	}

	taskRoute := RouteManager{
		Route:   route,
		Methods: []MethodHandler{testGetMethodHandler},
		Version: version,
	}
	return &taskRoute
}

type testGetHandler struct {
	taskId string
}

func (tgh *TestGetHandler) Handler() RequestHandler {
	return &testGethandler{}
}
func (tgh *testGetHandler) ParseAndValidate(r *http.Request) error {
	projCtx := MustHaveProjectContext(r)
	if projCtx.Task == nil {
		return apiv3.APIError{
			Message:    "Task not found",
			StatusCode: http.StatusNotFound,
		}
	}
	trh.taskId = projCtx.Task.Id
	return nil
}

func (tgh *testGetHandler) Execute(sc servicecontext.ServiceContext) (ResponseData, error) {
	tests, err := sc.FindTestsByTaskId(tgh.taskId)

	models := make([]model.Model, len(tests))
	for ix, t := range tests {
		apiTest := model.APITest{}
		err := apiTest.BuildFromService(t)
		if err != nil {
			return []model.Model{}, err
		}
		// Put the model into the array
		models[ix] = &apiTest
	}
	return models, nil
}
