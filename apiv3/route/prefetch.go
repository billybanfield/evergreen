package route

import (
	"fmt"
	"github.com/evergreen/model"
	"github.com/evergreen/model/user"
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
}

func PrefetchProject(r *http.Request, sc servicecontext.ServiceContext) error {
	vars := mux.Vars(r)
	taskId := vars["task_id"]
	buildId := vars["build_id"]
	versionId := vars["version_id"]
	patchId := vars["patch_id"]
	projectId := vars["project_id"]
	ctx, err := model.LoadContext(taskId, buildId, versionId, patchId, projectId)
	if err != nil {
		// Some database lookup failed when fetching the data - log it
		return err
	}

	if ctx.ProjectRef != nil && ctx.ProjectRef.Private && GetUser(r) == nil {
		return apiv3.APIError{
			StatusCode: http.StatusNotFound,
			Message:    "Project Not Found",
		}
	}

	if ctx.Patch != nil && GetUser(r) == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}

	context.Set(r, RequestContext, &ctx)
}
