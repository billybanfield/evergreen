/*
package route

import (
	"github.com/evergreen-ci/evergreen/model/user"
	"github.com/gorilla/context"
)

// GetUser returns a user if one is attached to the request. Returns nil if the user is not logged
// in, assuming that the middleware to lookup user information is enabled on the request handler.
func GetUser(r *http.Request) *user.DBUser {
	if rv := context.Get(r, RequestUser); rv != nil {
		return rv.(*user.DBUser)
	}
	return nil
}
*/
