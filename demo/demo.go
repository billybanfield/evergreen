package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/evergreen-ci/evergreen/apiv3/model"
)

//const ApiBaseUrl = "http://evergreenapp-1.staging.build.10gen.cc:8080/rest/v2"
const ApiBaseUrl = "http://10.4.125.11:8080/rest/v2"
const limit = 10000000

var linkMatcher = regexp.MustCompile(`^\<(\S+)\>; rel=\"(\S+)\"`)

func main() {

	// Fetch the paramaters for the execution of this restart script
	proj := flag.String("project", "mci", "the project to restart in")
	revision := flag.String("revision", "cbb4eb409e60df316e65fc8bbd4b59b5dcad9bc8", "the revision of the commit to restart")
	status := flag.String("status", "failed", "the status to restart with")
	flag.Parse()

	fmt.Printf("restarting tasks in project %s with revision %s and status %s\n", *proj, *revision, *status)

	// Fetch the tasks from the API
	taskIds, err := fetchTasksWithStatus(*proj, *revision, *status)
	if err != nil {
		panic(err)
	}

	// Call the API to restart the tasks
	if len(taskIds) > 0 {
		fmt.Println("found tasks to restart ")
		fmt.Printf("%v\n", taskIds)
		err = restartAllTasks(taskIds)
	} else {
		fmt.Println("no tasks found to restart")
	}

	if err != nil {
		panic(err)
	}

}

func fetchTasksWithStatus(project, version, status string) ([]string, error) {
	nextLink := fmt.Sprintf("%s/projects/%s/revisions/%s/tasks?limit=%d", ApiBaseUrl, project, version, limit)
	taskIds := []string{}
	var err error
	var tasks []model.APITask

	for nextLink != "" {
		tasks, nextLink, err = fetchPage(nextLink)
		if err != nil {
			return []string{}, err
		}
		for _, t := range tasks {
			if string(t.Status) == status {
				taskIds = append(taskIds, string(t.Id))
			}
		}

	}
	return taskIds, nil
}

func fetchPage(url string) ([]model.APITask, string, error) {
	apiKey := os.Getenv("API_KEY")

	// Create HTTP Request
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []model.APITask{}, "", fmt.Errorf("making request %v", err)
	}
	req.Header.Add("Api-User", "william.banfield")
	req.Header.Add("Api-Key", apiKey)

	// Perform API Request
	resp, err := client.Do(req)
	if err != nil {
		return []model.APITask{}, "", fmt.Errorf("performing request %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return []model.APITask{}, "", fmt.Errorf("server returned status %s", resp.Status)
	}

	// Fetch Result from Response
	tasks := []model.APITask{}
	err = json.NewDecoder(resp.Body).Decode(&tasks)
	if err != nil {
		return []model.APITask{}, "", fmt.Errorf("reading response body %v", err)
	}

	// Fetch Pagination Information
	links := resp.Header.Get("Link")
	matches := linkMatcher.FindStringSubmatch(links)
	fmt.Println(links)

	if len(matches) != 0 && len(matches) != 3 {
		return []model.APITask{}, "", fmt.Errorf("malformed link header %v", links)
	}
	var nextLink string
	if len(matches) != 0 {
		nextLink = matches[1]
	}

	return tasks, nextLink, nil
}

func restartAllTasks(taskIds []string) error {
	fmt.Println("restarting tasks")
	apiKey := os.Getenv("API_KEY")

	for _, t := range taskIds {
		// Create HTTP Request
		url := fmt.Sprintf("%s/tasks/%s/restart", ApiBaseUrl, t)
		client := &http.Client{}
		req, err := http.NewRequest("POST", url, nil)
		if err != nil {
			return fmt.Errorf("making request %v", err)
		}
		req.Header.Add("Api-User", "william.banfield")
		req.Header.Add("Api-Key", apiKey)

		// Perform API Request
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("performing request %v", err)
		}

		fmt.Println("server responded with status", resp.Status)
		task := model.APITask{}
		err = json.NewDecoder(resp.Body).Decode(&task)
		if err != nil {
			resp.Body.Close()
			return fmt.Errorf("reading response body %v", err)
		}
		fmt.Println(task)
	}
	return nil
}
