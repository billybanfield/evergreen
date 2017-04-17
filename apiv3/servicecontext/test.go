package servicecontext

// DBTestConnector is a struct that implements the Test related methods
// from the ServiceContext through interactions with the backing database.
type DBTestConnector struct{}

// MockTaskConnector stores a cached set of tests that are queried against by the
// implementations of the ServiceContext interface's Test related functions.
type MockTestConnector struct {
	CachedTests []task.Task
	StoredError error
}

func (tc *TaskConnector) FindTestsByTaskId(testId string) ([]task.TestResult, error) {
	t, err := task.FindOne(task.ById(taskId))
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, &apiv3.APIError{
			StatusCode: http.StatusNotFound,
			Message:    fmt.Sprintf("task with id %s not found", taskId),
		}
	}
	return t.TestResults, nil
}

func (mtc *MockTestConnector) FindTestsByTaskId(testId string) ([]task.TestResult, error) {
	return mtc.CachedTests, mtc.StoredError
}
