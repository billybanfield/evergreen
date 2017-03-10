package servicecontext

// DBServiceContext is a struct that implements all of the methods which
// connect to the service layer of evergreen. These methods abstract the link
// between the service and the API layers, allowing for changes in the
// service architecture without forcing changes to the API.
type DBServiceContext struct {
	superUsers []string

	DBUserConnector
	DBTaskConnector
	DBContextConnector
}

func (ctx *DBServiceContext) GetSuperUsers() []string {
	return ctx.superUsers
}

func (ctx *DBServiceContext) SetSuperUsers(su []string) {
	ctx.superUsers = su
}

type MockServiceContext struct {
	superUsers []string

	MockUserConnector
	MockTaskConnector
	MockContextConnector
}

func (ctx *MockServiceContext) GetSuperUsers() []string {
	return ctx.superUsers
}
func (ctx *MockServiceContext) SetSuperUsers(su []string) {
	ctx.superUsers = su
}
