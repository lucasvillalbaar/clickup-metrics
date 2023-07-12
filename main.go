package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
)

const (
	StatusToDo            = "to do"
	StatusInDefinitionPM  = "in definition pm"
	StatusInDefinitionDev = "in definition dev"
	StatusToDevelop       = "to develop"
	StatusInDevelopment   = "in development"
	StatusBlocked         = "blocked"
	StatusReadyToDeploy   = "ready to deploy"
	StatusComplete        = "completado"
)

// Structures to parser JSON
type Transition struct {
	Before TransitionState `json:"before"`
	After  TransitionState `json:"after"`
	Date   string          `json:"date"`
}

type TaskInfo struct {
	TaskId         string       `json:"task_id"`
	TaskDecription string       `json:"task_description"`
	TaskType       string       `json:"task_type"`
	StartDate      string       `json:"start_date"`
	History        []Transition `json:"history"`
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

func getTaskInfo(taskId string) TaskInfo {
	inputData := `
	{
		"history": [
			{
				"before": {
					"status": "in definition pm"
				},
				"after": {
					"status": "in definition dev"
				},
				"date": "1686578015783"
			},
			{
				"before": {
					"status": "in definition dev"
				},
				"after": {
					"status": "in development"
				},
				"date": "1687785838064"
			},
			{
				"before": {
					"status": "in development"
				},
				"after": {
					"status": "blocked"
				},
				"date": "1687995461000"
			},
			{
				"before": {
					"status": "blocked"
				},
				"after": {
					"status": "in development"
				},
				"date": "1688168261000"
			},
			{
				"before": {
					"status": "in development"
				},
				"after": {
					"status": "completado"
				},
				"date": "1688649601245"
			}
		]
	}`

	history := parserJSON([]byte(inputData))
	return TaskInfo{
		TaskId:    taskId,
		StartDate: "1685749061000",
		History:   history,
	}
}

func parserJSON(inputData []byte) []Transition {
	var taskInfo TaskInfo
	err := json.Unmarshal(inputData, &taskInfo)
	if err != nil {
		log.Fatal(err)
	}
	return taskInfo.History
}

func initMetrics() MetricsPerState {
	return MetricsPerState{
		StatusToDo:            {},
		StatusInDefinitionPM:  {},
		StatusInDefinitionDev: {},
		StatusToDevelop:       {},
		StatusInDevelopment:   {},
		StatusBlocked:         {},
		StatusReadyToDeploy:   {},
		StatusComplete:        {},
	}
}

func calculateTimePerState(taskInfo *TaskInfo) *MetricsPerState {
	metrics := initMetrics()
	//Set the start date of the first transition to the same value as the task start date
	metrics[Status(taskInfo.History[0].Before.Status)].StartDate = taskInfo.StartDate
	for _, entry := range taskInfo.History {
		//Complete Date for the new status
		statusAfter := Status(entry.After.Status)
		metricForAfterStatus := metrics[statusAfter]
		metricForAfterStatus.StartDate = entry.Date
		metrics[statusAfter] = metricForAfterStatus

		//Complete Date for previous status
		statusBefore := Status(entry.Before.Status)
		metricForBeforeStatus := metrics[statusBefore]
		metricForBeforeStatus.DueDate = entry.Date
		metrics[statusBefore] = metricForBeforeStatus

		metricForBeforeStatus.TimeSpent += calcTimeSpent(metricForBeforeStatus)
		fmt.Print("Estado: ", statusBefore, " Start Date: ", metricForBeforeStatus.StartDate, "Due Date: ", metricForBeforeStatus.DueDate)
		cleanDates(metricForBeforeStatus)
		fmt.Println("Time Spent: ", metricForBeforeStatus.TimeSpent)
	}

	return &metrics
}
func cleanDates(time *TimeSpent) {
	if time.StartDate != "" && time.DueDate != "" {
		time.StartDate = ""
		time.DueDate = ""
	}
}

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

	//Calculate Flow Efficiency
	metrics.FlowEfficiency = (float64(metrics.CycleTime) - float64(metrics.BlockedTime)) * 100 / float64(metrics.CycleTime)
	return &metrics
}

func isValidStateForCycleTime(status Status) bool {
	if status == StatusInDevelopment || status == StatusReadyToDeploy || status == StatusBlocked {
		return true
	}
	return false
}

func isStateWaitingOrBlocked(status Status) bool {
	return status == StatusBlocked
}

func main() {
	taskInfo := getTaskInfo("85zt8cyjd")
	metricsPerState := calculateTimePerState(&taskInfo)
	taskMetrics := calculateMetrics(metricsPerState)

	fmt.Printf("Lead Time: %d | Cycle Time: %d | Flow Efficiency: %.2f\n", taskMetrics.LeadTime, taskMetrics.CycleTime, taskMetrics.FlowEfficiency)
}
