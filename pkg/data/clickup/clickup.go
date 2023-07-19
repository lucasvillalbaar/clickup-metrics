package clickup

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/lucasvillalbaar/clickup-metrics/pkg/data"
)

// Constants for task statuses
const (
	StatusToDo            = "to do"
	StatusInDefinitionPM  = "in definition pm"
	StatusInDesign        = "in design"
	StatusInDefinitionDev = "in definition dev"
	StatusToDevelop       = "to develop"
	StatusInDevelopment   = "in development"
	StatusBlocked         = "blocked"
	StatusReadyToDeploy   = "ready to deploy"
	StatusDeployed        = "deployed"
	StatusComplete        = "completado"
)

const (
	EndpointTaskHistory = "https://app.clickup.com/tasks/v1/task/%s/history?reverse=true&hist_fields%%5B%%5D=status"
	EndpointTaskInfo    = "https://api.clickup.com/api/v2/task/%s"
)

type ResponseGetTask struct {
	Id        string `json:"id"`
	CustomId  string `json:"custom_id"`
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
	DueDate   string `json:"due_date"`
}

// getTaskHistory retrieves the task history from the ClickUp API
func getTaskHistory(taskId string) ([]byte, error) {
	url := fmt.Sprintf(EndpointTaskHistory, taskId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %v", err)
	}

	req.Header.Set("Authorization", s.token)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("error performing HTTP request: %v", err)
	}

	if resp.StatusCode == 401 {
		return nil, errors.New("token is expired")
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	return body, nil
}

// getTaskHeaderData retrieves the task header data from the ClickUp API
func getTaskHeaderData(taskId string) (data.TaskHeaderData, error) {
	url := fmt.Sprintf(EndpointTaskInfo, taskId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return data.TaskHeaderData{}, fmt.Errorf("error creating HTTP request: %v", err)
	}

	req.Header.Set("Authorization", s.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return data.TaskHeaderData{}, fmt.Errorf("error performing HTTP request: %v", err)
	}

	if resp.StatusCode == 401 {
		return data.TaskHeaderData{}, errors.New("api key is expired or is not valid")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return data.TaskHeaderData{}, fmt.Errorf("error reading response body: %v", err)
	}

	var response ResponseGetTask

	err = json.Unmarshal(body, &response)
	if err != nil {
		return data.TaskHeaderData{}, fmt.Errorf("error parsing body: %v", err)
	}

	return data.TaskHeaderData{
		Id:        response.Id,
		CustomId:  response.CustomId,
		Name:      response.Name,
		StartDate: response.StartDate,
		DueDate:   response.DueDate,
	}, nil
}

// parserJSON parses the JSON input into a slice of Transition
func parserJSON(inputData []byte) []data.Transition {
	var taskInfo data.TaskInfo
	err := json.Unmarshal(inputData, &taskInfo)
	if err != nil {
		log.Fatal(err)
	}
	return taskInfo.History
}

// getTaskInfo retrieves the task information including history and header data
func getTaskInfo(taskId string) (*data.TaskInfo, error) {
	historyData, err := getTaskHistory(taskId)
	if err != nil {
		return &data.TaskInfo{}, err
	}
	taskHeaderData, err := getTaskHeaderData(taskId)

	if err != nil {
		return &data.TaskInfo{}, err
	}

	history := parserJSON([]byte(historyData))

	return &data.TaskInfo{
		TaskHeaderData: taskHeaderData,
		History:        history,
	}, nil
}

type Session struct {
	apiKey string
	token  string
}

var s *Session

func Init(apiKey string, token string) *Session {
	s = &Session{
		apiKey: apiKey,
		token:  token,
	}

	return s
}

func (s *Session) GetTasksWithFilter(filter data.Filter) []data.TaskHeaderData {
	return nil
}

func (s *Session) GetTaskByID(id string) (*data.TaskInfo, error) {
	return getTaskInfo(id)
}

func (s *Session) GetHistoryPerTask(ids []string) *map[string]data.TaskInfo {
	return nil
}

func (s *Session) GetWorkflow() *data.Workflow {
	var statuses []data.Status
	statuses = append(statuses, data.Status{
		Name:       StatusToDo,
		Pending:    false,
		InProgress: false,
		Blocked:    false,
		Done:       false,
	}, data.Status{
		Name:       StatusInDefinitionPM,
		Pending:    false,
		InProgress: true,
		Blocked:    false,
		Done:       false,
	}, data.Status{
		Name:       StatusInDesign,
		Pending:    false,
		InProgress: true,
		Blocked:    false,
		Done:       false,
	}, data.Status{
		Name:       StatusInDefinitionDev,
		Pending:    false,
		InProgress: true,
		Blocked:    false,
		Done:       false,
	}, data.Status{
		Name:       StatusToDevelop,
		Pending:    true,
		InProgress: false,
		Blocked:    false,
		Done:       false,
	}, data.Status{
		Name:       StatusInDevelopment,
		Pending:    false,
		InProgress: true,
		Blocked:    false,
		Done:       false,
	}, data.Status{
		Name:       StatusBlocked,
		Pending:    false,
		InProgress: false,
		Blocked:    true,
		Done:       false,
	}, data.Status{
		Name:       StatusReadyToDeploy,
		Pending:    true,
		InProgress: false,
		Blocked:    false,
		Done:       false,
	}, data.Status{
		Name:       StatusDeployed,
		Pending:    false,
		InProgress: false,
		Blocked:    false,
		Done:       true,
	}, data.Status{
		Name:       StatusComplete,
		Pending:    false,
		InProgress: false,
		Blocked:    false,
		Done:       true,
	})

	return &data.Workflow{
		Statuses: statuses,
	}
}
