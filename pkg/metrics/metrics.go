package metrics

import (
	"fmt"
	"log"
	"strconv"
)

type Metrics struct {
	LeadTime       int     `json:"lead_time"`
	CycleTime      int     `json:"cycle_time"`
	BlockedTime    int     `json:"blocked_time"`
	FlowEfficiency float64 `json:"flow_efficiency"`
}

type TimeSpent struct {
	StartDate string `json:"start_date"`
	DueDate   string `json:"due_date"`
	TimeSpent int    `json:"time_spent"`
}

type MetricsPerState map[string]*TimeSpent

type MetricsPerTask struct {
	TaskInfo TaskInfo `json:"task_info"`
	Metrics  Metrics  `json:"metrics"`
}

type Status struct {
	Name                  string
	Pending               bool
	InProgress            bool
	Blocked               bool
	Done                  bool
	IsLeadTimeCalculable  bool
	IsCycleTimeCalculable bool
}

type Workflow struct {
	Statuses map[string]Status
}

type Transition struct {
	Before string `json:"before"` // Old Status
	After  string `json:"after"`  // New Status
	Date   string `json:"date"`   // Change Date
}

type TaskInfo struct {
	Id        string       // Task ID
	Name      string       // Task name
	StartDate string       // Start date of the task
	DueDate   string       // Due Date of the task
	History   []Transition // All history data of the task
}

var wf Workflow
var metrics MetricsPerState
var taskInfo []TaskInfo

// initMetrics initializes the MetricsPerState map with empty values for each status
func InitMetrics(workflow Workflow) {
	wf = workflow
	metrics = MetricsPerState{}

	if len(workflow.Statuses) == 0 {
		log.Fatal("Metrics: Workflow is empty")
	}
	for _, entry := range workflow.Statuses {
		metrics[entry.Name] = &TimeSpent{}
	}
}

func SetInfo(tasks []TaskInfo) {
	taskInfo = tasks
}

// calculateTimePerState calculates the time spent in each state for the task
func calculateTimePerState(taskInfo TaskInfo) *MetricsPerState {
	// Set the start date of the first transition to the same value as the task start date
	metrics[taskInfo.History[0].Before].StartDate = taskInfo.StartDate
	for _, entry := range taskInfo.History {
		if entry.Date < taskInfo.StartDate {
			continue
		}
		// Complete Date for the new status
		statusAfter := entry.After
		metricForAfterStatus := metrics[statusAfter]
		metricForAfterStatus.StartDate = entry.Date
		metrics[statusAfter] = metricForAfterStatus

		// Complete Date for previous status
		statusBefore := entry.Before
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
func CalculateMetrics() []MetricsPerTask {
	metricsPerTask := []MetricsPerTask{}

	for _, ti := range taskInfo {
		mpt := calculateTimePerState(ti)
		metrics := MetricsPerTask{
			TaskInfo: ti,
		}
		for status, entry := range *mpt {
			metrics.Metrics.LeadTime += entry.TimeSpent
			if wf.Statuses[status].IsCycleTimeCalculable {
				metrics.Metrics.CycleTime += entry.TimeSpent
			}
			if wf.Statuses[status].Blocked || wf.Statuses[status].Pending {
				metrics.Metrics.BlockedTime += entry.TimeSpent
			}
		}
		// Calculate Flow Efficiency
		metrics.Metrics.FlowEfficiency = (float64(metrics.Metrics.CycleTime) - float64(metrics.Metrics.BlockedTime)) * 100 / float64(metrics.Metrics.CycleTime)
		metricsPerTask = append(metricsPerTask, metrics)
	}

	return metricsPerTask
}
