package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/lucasvillalbaar/clickup-metrics/pkg/configuration"
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

// Structures to parse JSON
type Transition struct {
	Before TransitionState `json:"before"`
	After  TransitionState `json:"after"`
	Date   string          `json:"date"`
}

type TaskHeaderData struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
}

type TaskInfo struct {
	TaskHeaderData
	History []Transition `json:"history"`
}

type TransitionState struct {
	Status string `json:"status"`
}

type Metrics struct {
	LeadTime       int     `json:"lead_time"`
	CycleTime      int     `json:"cycle_time"`
	BlockedTime    int     `json:"blocked_time"`
	FlowEfficiency float64 `json:"flow_efficiency"`
}

type Status string

type TimeSpent struct {
	StartDate string `json:"start_date"`
	DueDate   string `json:"due_date"`
	TimeSpent int    `json:"time_spent"`
}

type MetricsPerState map[Status]*TimeSpent

type TasksMetrics struct {
	TaskId  string  `json:"task_id"`
	Metrics Metrics `json:"metrics"`
}

// getTaskHistory retrieves the task history from the ClickUp API
func getTaskHistory(taskId string) ([]byte, error) {
	url := fmt.Sprintf("https://app.clickup.com/tasks/v1/task/%s/history?reverse=true&hist_fields%%5B%%5D=status", taskId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+configuration.GetEnvironmentVariables().Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error performing HTTP request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	return body, nil
}

// getTaskHeaderData retrieves the task header data from the ClickUp API
func getTaskHeaderData(taskId string) (TaskHeaderData, error) {
	url := fmt.Sprintf("https://api.clickup.com/api/v2/task/%s", taskId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return TaskHeaderData{}, fmt.Errorf("error creating HTTP request: %v", err)
	}

	req.Header.Set("Authorization", configuration.GetEnvironmentVariables().ApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return TaskHeaderData{}, fmt.Errorf("error performing HTTP request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return TaskHeaderData{}, fmt.Errorf("error reading response body: %v", err)
	}

	var taskHeaderData TaskHeaderData

	err = json.Unmarshal(body, &taskHeaderData)
	if err != nil {
		return TaskHeaderData{}, fmt.Errorf("error parsing body: %v", err)
	}

	return taskHeaderData, nil
}

// getTaskInfo retrieves the task information including history and header data
func getTaskInfo(taskId string) TaskInfo {
	historyData, _ := getTaskHistory(taskId)
	taskHeaderData, _ := getTaskHeaderData(taskId)

	history := parserJSON([]byte(historyData))

	return TaskInfo{
		TaskHeaderData: taskHeaderData,
		History:        history,
	}
}

// parserJSON parses the JSON input into a slice of Transition
func parserJSON(inputData []byte) []Transition {
	var taskInfo TaskInfo
	err := json.Unmarshal(inputData, &taskInfo)
	if err != nil {
		log.Fatal(err)
	}
	return taskInfo.History
}

// initMetrics initializes the MetricsPerState map with empty values for each status
func initMetrics() MetricsPerState {
	return MetricsPerState{
		StatusToDo:            {},
		StatusInDefinitionPM:  {},
		StatusInDesign:        {},
		StatusInDefinitionDev: {},
		StatusToDevelop:       {},
		StatusInDevelopment:   {},
		StatusBlocked:         {},
		StatusReadyToDeploy:   {},
		StatusDeployed:        {},
		StatusComplete:        {},
	}
}

// calculateTimePerState calculates the time spent in each state for the task
func calculateTimePerState(taskInfo *TaskInfo) *MetricsPerState {
	metrics := initMetrics()
	// Set the start date of the first transition to the same value as the task start date
	metrics[Status(taskInfo.History[0].Before.Status)].StartDate = taskInfo.StartDate
	for _, entry := range taskInfo.History {
		// Complete Date for the new status
		statusAfter := Status(entry.After.Status)
		metricForAfterStatus := metrics[statusAfter]
		metricForAfterStatus.StartDate = entry.Date
		metrics[statusAfter] = metricForAfterStatus

		// Complete Date for previous status
		statusBefore := Status(entry.Before.Status)
		metricForBeforeStatus := metrics[statusBefore]
		metricForBeforeStatus.DueDate = entry.Date
		metrics[statusBefore] = metricForBeforeStatus

		metricForBeforeStatus.TimeSpent += calcTimeSpent(metricForBeforeStatus)
		cleanDates(metricForBeforeStatus)
	}

	return &metrics
}

// cleanDates sets StartDate and DueDate to empty if both are present
func cleanDates(time *TimeSpent) {
	if time.StartDate != "" && time.DueDate != "" {
		time.StartDate = ""
		time.DueDate = ""
	}
}

// calcTimeSpent calculates the time spent between StartDate and DueDate in days
func calcTimeSpent(time *TimeSpent) int {
	if time.StartDate == "" || time.DueDate == "" {
		return 0
	}

	// Dates in Unix format in milliseconds
	startDate := time.StartDate
	endDate := time.DueDate

	// Convert the dates in milliseconds to integers
	startTime, err := strconv.ParseInt(startDate, 10, 64)
	if err != nil {
		fmt.Println("Error converting start date:", err)
	}

	endTime, err := strconv.ParseInt(endDate, 10, 64)
	if err != nil {
		fmt.Println("Error converting end date:", err)
	}

	// Calculate the difference in days
	difference := int((endTime - startTime) / (1000 * 60 * 60 * 24))

	// Round up if the difference is more than half a day
	if (endTime-startTime)%(1000*60*60*24) >= (1000*60*60*24)/2 {
		difference++
	}

	return difference
}

// calculateMetrics calculates the overall metrics based on the time spent in each state
func calculateMetrics(metricsPerState *MetricsPerState) *Metrics {
	var metrics Metrics

	for status, entry := range *metricsPerState {
		metrics.LeadTime += entry.TimeSpent
		if isValidStateForCycleTime(status) {
			metrics.CycleTime += entry.TimeSpent
		}
		if isStateWaitingOrBlocked(status) {
			metrics.BlockedTime += entry.TimeSpent
		}
	}

	// Calculate Flow Efficiency
	metrics.FlowEfficiency = (float64(metrics.CycleTime) - float64(metrics.BlockedTime)) * 100 / float64(metrics.CycleTime)
	return &metrics
}

// isValidStateForCycleTime checks if a status is valid for calculating CycleTime
func isValidStateForCycleTime(status Status) bool {
	if status == StatusInDevelopment || status == StatusReadyToDeploy || status == StatusBlocked {
		return true
	}
	return false
}

// isStateWaitingOrBlocked checks if a status indicates a waiting or blocked state
func isStateWaitingOrBlocked(status Status) bool {
	return (status == StatusBlocked) || (status == StatusToDevelop)
}

func main() {
	err := configuration.LoadEnvironmentVariables()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	taskInfo := getTaskInfo("85zrzu15w")
	metricsPerState := calculateTimePerState(&taskInfo)
	taskMetrics := calculateMetrics(metricsPerState)

	fmt.Printf("Task ID: %s | %s | %s\n", taskInfo.Id, taskInfo.Name, taskInfo.StartDate)
	fmt.Printf("Lead Time: %d | Cycle Time: %d | Flow Efficiency: %.2f\n", taskMetrics.LeadTime, taskMetrics.CycleTime, taskMetrics.FlowEfficiency)
}
