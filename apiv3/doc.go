/*
Developing for the REST V2 API

Introduction to Types and their Uses

Model

Models are structs which represent the the object returned by the API. A 'model'
is an interface with two methods, BuildFromService and ToService which
define how to transform to and from an API model.

ServiceContext

ServiceContext is a very large interface that defines interaction with the backing
database and service layer. It has two main sets of implementations: one set that
communicate with the database and one that mocks the same functionality.

RequestHandler

RequestHandler is an interface which defines how to process an HTTP request
for an API resource. RequestHandlers implement methods for fetching a new copy
of the RequestHandler, ParseAndValidate the request body and how to Execute
the bulk of the request against the database.

MethodHandler

MethodHandler is an struct that contains all of the data necessary for
completing an API request. It contains an authenticator to handle much of the
access control, functions for prefetching data and attaching it to requests, and
a RequestHandler to execute the request.

RouteManager

RouteManagers are structs that define all of the functionality of a particular
API route, including definitions of each of the methods it implements and the
path used to access it.

PaginationExecutor

PaginationExecutor is a type that handles gathering necessary information for
paginating and handles executing the necessary parts of the API request. The
PaginationExecutor type implements the RequestHandler so that writing paginated
endpoints does not necessarily require rewriting these methods; However, any of
these methods may be overwritten to provide additional flexibility and funcitonality.

PaginatorFunc

PaginatorFunc is a function type that defines how to perform pagination over a
specific resource. It is the only type that is required to be implemented when
adding paginated API resources.

Adding a Route

Adding a new route to the REST v2 API requires creation of a few structs and
implementation of a few new methods.

RequestHandler

The RequestHandler is a central interface in the REST v2 API with following
signature:


				type RequestHandler interface {
					Handler() RequestHandler
					ParseAndValidate(*http.Request) error
					Execute(servicecontext.ServiceContext) (ResponseData, error)
				}

RequestHandlers are placed in files in the route/ directory depending on the type
of resource they access.

To add a new route you must create a struct that implements its three main interface
methods. The Handler method must return a new copy of the RequestHandler so that
a new copy of this object can be used on successive calls to the same endpoint.

The ParseAndValidate method is the only method that has access to the http.Request object.
All necessary query parameters and request body information must be fetched from the request
in this function and added to the struct for use in the Execute function. These fetches can
take a few main forms:

From mux Context
 Data gathered before the main request by the PrefetchFunc's are attached to the
mux Context for that request and can be fetched using the context.Get function
and providing it with the correct key for the desired data.

From the Route Variables
 Variables from routes defined with variables such as /tasks/{task_id} can be
fetched using calls to the mux.Vars funciton and providing the variable name
to the returned map. For example, the taskId of that route could be fetched using:

				mux.Vars(r)["task_id"]

From the URL Query Parameters
To fetch variables from the URL query parameters, get it from the http.Request's
URL object using:

				r.URL.Query().Get("status")

Finally, the Execute method is the only method with access to the ServiceContext
and is therefore capable of making calls to the backing database to fetch and alter
its state. The Execute method should use the parameters gathered in the ParseAndValidate
method to implement the main logic and functionality of the request.

Pagination

PaginationExecutor is a struct that already implements the RequestHandler interface.
To create a method with pagination, the only function that is needed is a PaginatorFunc.

PaginatorFunc

A PaginatorFunc defines how to paginate over a resource given a key to start pagination
from and a limit to limit the number of results. PaginatorFunc has the following signature:


				func(key string, limit int, args interface{}, sc ServiceContext)([]Model, *PageResult, error)

The key and limit are fetched automatically by the PaginationExecutor's ParseAndValidate
function. These parameters should be used to query for the correct set of results.

The args is a parameter that may optionally be used when more information is
needed to completed the request. To populate this field, the RequestHandler that
wraps the PaginationExecutor must implement a ParseAndValidate method that overwrites
the PaginationExecutor's and then calls it with the resulting request for example,
a RequestHandler called fooRequestHandler that needs additional args would look
like:

				fooRequestHandler{
				 *PaginationExecutor
				}

				extraFooArgs{
					extraParam string
				}

				func(f *fooRequesetHandler) ParseAndValidate(r *http.RequestHandler) error{
					urlParam := r.URL.Query().Get("extra_url_param")
					f.PaginationExecutor.Args = extraFooArgs{urlParam}

					return f.PaginationExecutor.ParseAndValidate(r)
				}

				func fooRequestPaginator(key string, limit int, args interface{},
					 sc servicecontext.ServiceContext)([]model.Model, *PageResult, error){

					 fooArgs, ok := args.(extraFooArgs)
					 if !ok {
						// Error
					 }

				...
				}


PageResult

The PageResult is a struct that must be constructed and returned by a PaginatorFunc
It contains the information used for creating links to the next and previous page of
results.

To construct a Page, you must provide it with the limit of the number of results
for the page, which is either the default limit if none was provided, the limit
of the previous request if provided, or the number of documents between the page
and the end of the result set. The end of the result set is either the beginning of
set of results currently being returned if constructing a previous page, or the end
of all results if constructing a next page.

The Page must also contain the key of the item that begins the Page of results.
For example, when creating a next Page when fetching a page of 100 tasks, the
task_id of the 101st task should be used as the key for the next Page.

If the page being returned is the first or last page of pagination, then there
is no need to create that Page.

MethodHandler

The MethodHandler type contains all data and types for complete execution of an
API method. It holds:
A list of PrefetchFuncs used for grabbing data needed before
the main execution of a method, such as user data for authentication

The HTTP method type (GET, PUT, etc.)

The Authenticator this method uses to control access to its data

The RequestHandler this method uses for the main execution of its request

A MethodHandler is a compositon of defined structs and functions that in total
comprise the method. Many Authenticator and PrefetchFunc are already implemented
and only need to be attached to this object to create the method once the Requesthandler
is complete.

RouteManager

The RouteManager type holds all of the methods associated with a particular API
route. It holds these as an array of MethodHandlers. It also contains the path
by which this route may be accessed, and the version of the API that this is
implemented as part of.

When adding to the API there may already be a RouteManger in existence for the
method being devloped. For example, if a method for GET /tasks/<task_id> is
already implemented, then a new route manager is not required when createing POST /tasks/<task_id>.
Its implementation only needs to be added to the existing RouteManager.

Once created, the RouteManager must be registered onto the mux.Router with the
ServiceContext by calling route.Register(router, serviceContext).

Adding Models

Each model is kept in the model package of the REST v2 API in its own file.
To create a new model, define a struct containing all of the fields that it will return
and implement its two main interface methods BuildFromService and ToService.
Be sure to include struct tags to the define the names the fields will have when
serialized to JSON.

Guidlines for Creating Models

Include as much data as a user is likely to want when inspecting this resource.
This is likely to be more information than seems directly needed, but there is
little penalty to its inclusion.

Use APIString instead of Golang's string type. APIString serializes emtpy strings
as JSON null instead of Go's zero type of '""'

Use APITime instead of go's time type. APITime is a type that wraps Go's time.Time
and automatically correctly serializes it to ISO-8601 UTC time

Return an error when TypeCasting fails.

Model Methods

				BuildFromService(in interface{}) error

BuildFromService fetches all needed data from the passed in object and sets them
on the model. BuildFromService may sometimes be called multiple times with different
types that all contain data to build up the model object. In this case, a type switch
is likely necessary to determine what has been passed in.

				ToService()(interface{}, error)

ToService creates an as-complete-as-possible version of the service layer's version
of this model. For example, if this is is a REST v2 Task model, the ToService method
creates a service layer Task and sets all of the fields it is able to and returns it.


Adding to the ServiceContext

The ServiceContext is a very large interface that defines how to access the main
state of the database and data central to Evergreen's function. All methods of the
ServiceContext are contained in the servicecontext package in files depending
on the main type they allow access to (i.e. all test access is contained in servicecontext/test.go).

ServiceContext should only be done when the desired functionality cannot be performed
using a combination of the methods it already contains OR when such combination would
be unseemingly slow or expensive.

To add to the ServiceContext, add the method signature into the interface in
servicecontext/servicecontext.go. Next, add the implementation that interacts
with the database to the database backed object. These objects are named by the
resource they allow access to. The object that allows access to Hosts is called
DBHostConnector. Finally, add a mock implementation to the mock object. For
Hosts again, this object would be called MockHostConnector.

Implementing database backed methods requires using methods in Evergreen's model
package. As much database specific information as possible should be kept out of
the these methods. For example, if a new aggregation pipeline is needed to complete
the request, it should be defined int the Evergreen model package and used only
to aggregate in the method.
*/

package apiv3
