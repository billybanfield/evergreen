package servicecontext

import (
	"github.com/evergreen-ci/evergreen/model/user"
)

// UserConnector is an interface which abstracts the service layer's fetching
// and modifactions of users.
type UserConnector interface {

	// FindUserById is a method to find a specific user given its ID.
	FindUserById(string) (*user.DBUser, error)
}

// DBUserConnector is a struct that implements the UserConnector interface
// through interactions with the backing database.
type DBUserConnector struct{}

// FindUserById uses the service layer's user type to query the backing database for
// the user with the given Id.
func (tc *DBUserConnector) FindUserById(userId string) (*user.DBUser, error) {
	t, err := user.FindOne(user.ById(userId))
	if err != nil {
		return nil, err
	}
	return t, nil
}

// MockUserConnector stores a cached set of users that are queried against by the
// implementations of the UserConnector interface's functions.
type MockUserConnector struct {
	CachedUsers map[string]*user.DBUser
}

// FindUserById provides a mock implementation of a UserConnector that does not
// need to use a database. It returns results based on the cached users
// in the MockUserConnector.
func (muc *MockUserConnector) FindUserById(userId string) (*user.DBUser, error) {
	u := muc.CachedUsers[userId]
	return u, nil
}
