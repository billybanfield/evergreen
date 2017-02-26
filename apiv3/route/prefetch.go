package route

import (
	"fmt"
	"net/http"

	"github.com/evergreen-ci/evergreen/apiv3"
	"github.com/evergreen-ci/evergreen/apiv3/servicecontext"
	"github.com/evergreen-ci/evergreen/model"
	"github.com/evergreen-ci/evergreen/model/user"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

type (
	// custom types used to attach specific values to request contexts, to prevent collisions.
	requestUserKey    int
	requestContextKey int
)

const (
	// Key values used to map user and project data to request context.
	// These are private custom types to avoid key collisions.
	RequestUser    requestUserKey    = 0
	RequestContext requestContextKey = 0
)

type PrefetchFunc func(r *http.Request, sc servicecontext.ServiceContext) error

// UserMiddleware is middleware which checks for session tokens on the Request
// and looks up and attaches a user for that token if one is found.
func PrefetchUser(r *http.Request, sc servicecontext.ServiceContext) error {
	// Grab API auth details from header
	var authDataAPIKey, authDataName string

	if len(r.Header["Api-Key"]) > 0 {
		authDataAPIKey = r.Header["Api-Key"][0]
	}
	if len(r.Header["Auth-Username"]) > 0 {
		authDataName = r.Header["Auth-Username"][0]
	}
	if len(authDataName) == 0 && len(r.Header["Api-User"]) > 0 {
		authDataName = r.Header["Api-User"][0]
	}

	if len(authDataAPIKey) > 0 {
		dbUser, err := sc.FindUserById((authDataName))
		if dbUser != nil && err == nil {

			if dbUser.APIKey != authDataAPIKey {
				return apiv3.APIError{
					StatusCode: http.StatusUnauthorized,
					Message:    "Invalid API key",
				}
			}
			context.Set(r, RequestUser, dbUser)
		}
	}
	return nil
}

func PrefetchProjectContext(r *http.Request, sc servicecontext.ServiceContext) error {
	vars := mux.Vars(r)
	taskId := vars["task_id"]
	buildId := vars["build_id"]
	versionId := vars["version_id"]
	patchId := vars["patch_id"]
	projectId := vars["project_id"]
	ctx, err := sc.FetchContext(taskId, buildId, versionId, patchId, projectId)
	if err != nil {
		return err
	}

	if ctx.ProjectRef != nil && ctx.ProjectRef.Private && GetUser(r) == nil {
		// Project is private and user is not authorized so return not found
		return apiv3.APIError{
			StatusCode: http.StatusNotFound,
			Message:    "Project Not Found",
		}
	}

	if ctx.Patch != nil && GetUser(r) == nil {
		return apiv3.APIError{
			StatusCode: http.StatusNotFound,
			Message:    "Not Found",
		}
	}
	context.Set(r, RequestContext, &ctx)
	return nil
}

func GetUser(r *http.Request) *user.DBUser {
	if rv := context.Get(r, RequestUser); rv != nil {
		return rv.(*user.DBUser)
	}
	return nil
}

func GetProjectContext(r *http.Request) (*model.Context, error) {
	if rv := context.Get(r, RequestContext); rv != nil {
		return rv.(*model.Context), nil
	}
	return &model.Context{}, fmt.Errorf("No context loaded")
}

func MustHaveProjectContext(r *http.Request) *model.Context {
	pc, err := GetProjectContext(r)
	if err != nil {
		panic(err)
	}
	return pc
}
