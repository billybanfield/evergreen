package model

type APITest struct {
	Status    APIString `json:"status"`
	TestFile  APIString `json:"test_file"`
	Logs      testLog   `json:"logs"`
	ExitCode  int       `json:"exit_code"`
	StartTime APITime   `json:"start_time"`
	EndTime   APITime   `json:"end_time"`
}

type APITests struct {
	TaskId  APIString `json:"task_id"`
	TestMap `json:"test_results"`
}

type testLogs struct {
	URL    APIString `json:"url"`
	URLRaw APIString `json:"url_raw"`
}

type TestMap map[string]APITest

func (at *APITests) BuildFromService(st interface{}) error {
	switch v := st.(type) {
	case []task.TestResult:
		m := make(map[string]APITest, len(v))
		testMap := TestMap(m)
		for _, t := range v {
			apiTest := &APITest{}
			apiTest.BuildFromService(t)
			testMap[apiTest.TestFile] = apiTest
		}
		at.TestMap = testMap
	case string:
		at.TaskId = v
	default:
		return fmt.Errorf("Incorrect type when creating APITests")
	}
	return nil
}

func (at *APITests) ToService() (interface{}, error) {
	return nil, nil
}

func (at *APITest) BuildFromService(st interface{}) error {
	switch v := st.(type) {
	case task.TestResult:
		(*at) = APITest {
			Status: APIString(v.Status),
			TestFile: APIString(v.TestFile),
			ExitCode: v.ExitCode,
		}

		startTime := time.Time(v.StartTime)
		endTime := time.Time(v.EndTime)

		at.StartTime = APITime(startTime)
		at.EndTime = APITime(endTime)




		}

	default:
		return fmt.Errorf("Incorrect type when creating APITest")
	}
	return nil

}

func (at *APITest) ToService() (interface{}, error) {
	return nil, nil
}
