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
	StatusBacklog         = "backlog"
	StatusInDefinitionPM  = "in definition pm"
	StatusRefining        = "refining"
	StatusReadyForDev     = "ready for dev"
	StatusInDesign        = "in design"
	StatusInDefinitionDev = "in definition dev"
	StatusToDevelop       = "to develop"
	StatusInDevelopment   = "in development"
	StatusInTesting       = "in testing"
	StatusInValidation    = "in validation"
	StatusBlocked         = "blocked"
	StatusReadyToDeploy   = "ready to deploy"
	StatusDeployed        = "deployed"
	StatusCanceled        = "canceled"
	StatusCompletado      = "completado"
	StatusCompleted       = "completed"
	StatusReleased        = "released"
)

// Constants for task statuses (Data Team)
const (
	DataStatusToReview   = "to review"   // p90020097624_8pLbE48t
	DataStatusInProgress = "in progress" // p90020097624_GXrdPBxq
	DataStatusBlocked    = "blocked"     // p90020097624_YSfel0YN
	DataStatusClosed     = "Closed"      // p90020097624_itsGDThw
)

const (
	EndpointTaskHistory = "https://api.clickup.com/api/v2/task/%s/time_in_status"
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
func getTaskHistory(taskId string) ([]data.History, error) {
	url := fmt.Sprintf(EndpointTaskHistory, taskId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %v", err)
	}

	req.Header.Set("Authorization", s.apiKey)

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

	timeInStatus := data.TimeInStatusResponse{}

	err = json.Unmarshal(body, &timeInStatus)
	if err != nil {
		return nil, err
	}

	history := []data.History{}

	for _, status := range timeInStatus.StatusHistory {
		history = append(history, data.History{
			Status: status.Status,
			Time:   status.TotalTime.ByMinute,
			Since:  status.TotalTime.Since,
		})
	}

	return history, nil
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
	log.Println("Fetching data from Clickup for ticket:", taskId)
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

// getTaskInfo retrieves the task information including history and header data
func getTaskInfo(taskId string) (*data.TaskInfo, error) {
	history, err := getTaskHistory(taskId)
	if err != nil {
		return &data.TaskInfo{}, err
	}
	taskHeaderData, err := getTaskHeaderData(taskId)

	if err != nil {
		return &data.TaskInfo{}, err
	}

	return &data.TaskInfo{
		TaskHeaderData: taskHeaderData,
		History:        history,
	}, nil
}

type Session struct {
	apiKey string
}

var s *Session

func Init(apiKey string) *Session {
	s = &Session{
		apiKey: apiKey,
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
		Name:       StatusBacklog,
		Pending:    false,
		InProgress: false,
		Blocked:    false,
		Done:       false,
	}, data.Status{
		Name:       StatusRefining,
		Pending:    false,
		InProgress: true,
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
		Name:       StatusReadyForDev,
		Pending:    true,
		InProgress: false,
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
		Name:       StatusInTesting,
		Pending:    false,
		InProgress: true,
		Blocked:    false,
		Done:       false,
	}, data.Status{
		Name:       StatusInValidation,
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
		Name:       StatusReleased,
		Pending:    false,
		InProgress: false,
		Blocked:    false,
		Done:       true,
	}, data.Status{
		Name:       StatusCanceled,
		Pending:    false,
		InProgress: false,
		Blocked:    false,
		Done:       true,
	}, data.Status{
		Name:       StatusCompletado,
		Pending:    false,
		InProgress: false,
		Blocked:    false,
		Done:       true,
	}, data.Status{
		Name:       StatusCompleted,
		Pending:    false,
		InProgress: false,
		Blocked:    false,
		Done:       true,
	}, data.Status{
		Name:       DataStatusToReview,
		Pending:    false,
		InProgress: false,
		Blocked:    false,
		Done:       false,
	}, data.Status{
		Name:       DataStatusInProgress,
		Pending:    false,
		InProgress: true,
		Blocked:    false,
		Done:       false,
	}, data.Status{
		Name:       DataStatusBlocked,
		Pending:    false,
		InProgress: false,
		Blocked:    true,
		Done:       false,
	}, data.Status{
		Name:       DataStatusClosed,
		Pending:    false,
		InProgress: false,
		Blocked:    false,
		Done:       true,
	},
	)

	return &data.Workflow{
		Statuses: statuses,
	}
}
