package servicecontext

import (
	"github.com/evergreen-ci/evergreen/model/user"

	"github.com/gorilla/context"
)

const (
	// Key values used to map user and project data to request context.
	// These are private custom types to avoid key collisions.
	RequestUser reqUserKey = 0
)

type requestContext struct {
	Project *model.Context
}

// GetUser returns a user if one is attached to the request. Returns nil if the user is not logged
// in, assuming that the middleware to lookup user information is enabled on the request handler.
func GetUser(r *http.Request) *user.DBUser {
	if rv := context.Get(r, RequestUser); rv != nil {
		return rv.(*user.DBUser)
	}
	return nil
}

// UserMiddleware is middleware which checks for session tokens on the Request
// and looks up and attaches a user for that token if one is found.
func UserMiddleware(um auth.UserManager) func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		token := ""
		var err error
		// Grab token auth from cookies
		for _, cookie := range r.Cookies() {
			if cookie.Name == evergreen.AuthTokenCookie {
				if token, err = url.QueryUnescape(cookie.Value); err == nil {
					break
				}
			}
		}

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

		if len(token) > 0 {
			dbUser, err := um.GetUserByToken(token)
			if err != nil {
				evergreen.Logger.Logf(slogger.INFO, "Error getting user %v: %v", authDataName, err)
			} else {
				// Get the user's full details from the DB or create them if they don't exists
				dbUser, err := model.GetOrCreateUser(dbUser.Username(), dbUser.DisplayName(), dbUser.Email())
				if err != nil {
					evergreen.Logger.Logf(slogger.INFO, "Error looking up user %v: %v", dbUser.Username(), err)
				} else {
					context.Set(r, RequestUser, dbUser)
				}
			}
		} else if len(authDataAPIKey) > 0 {
			dbUser, err := user.FindOne(user.ById(authDataName))
			if dbUser != nil && err == nil {
				if dbUser.APIKey != authDataAPIKey {
					http.Error(rw, "Unauthorized - invalid API key", http.StatusUnauthorized)
					return
				}
				context.Set(r, RequestUser, dbUser)
			} else {
				evergreen.Logger.Logf(slogger.ERROR, "Error getting user: %v", err)
			}
		}
		next(rw, r)
	}
}
