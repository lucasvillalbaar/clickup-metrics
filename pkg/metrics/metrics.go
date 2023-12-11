package metrics

import (
	"log"
	"time"

	"github.com/lucasvillalbaar/clickup-metrics/pkg/data"
)

type Metrics struct {
	LeadTime       int     `json:"lead_time"`
	CycleTime      int     `json:"cycle_time"`
	BlockedTime    int     `json:"blocked_time"`
	FlowEfficiency float64 `json:"flow_efficiency"`
}

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
	Id        string         // Task ID
	Name      string         // Task name
	StartDate string         // Start date of the task
	DueDate   string         // Due Date of the task
	History   []data.History // All history data of the task
}

var wf Workflow
var taskInfo []TaskInfo

// initMetrics initializes the MetricsPerState map with empty values for each status
func InitMetrics(workflow Workflow) {
	wf = workflow
	if len(workflow.Statuses) == 0 {
		log.Fatal("Metrics: Workflow is empty")
	}
}

func SetInfo(tasks []TaskInfo) {
	taskInfo = tasks
}

func minutesToDays(minutes int) int {
	duration := time.Duration(minutes) * time.Minute
	days := int(duration.Hours() / 24)

	if duration.Hours() > float64(days*24) {
		days++
	}

	if days < 1 {
		return 1
	}

	return days
}

// calculateMetrics calculates the overall metrics based on the time spent in each state
func CalculateMetrics() []MetricsPerTask {
	metricsPerTask := []MetricsPerTask{}

	for _, ti := range taskInfo {
		metrics := MetricsPerTask{
			TaskInfo: ti,
		}
		for _, entry := range ti.History {
			metrics.Metrics.LeadTime += minutesToDays(entry.Time)
			if wf.Statuses[entry.Status].IsCycleTimeCalculable {
				metrics.Metrics.CycleTime += minutesToDays(entry.Time)
			}
			if wf.Statuses[entry.Status].Blocked || wf.Statuses[entry.Status].Pending {
				metrics.Metrics.BlockedTime += minutesToDays(entry.Time)
			}
		}

		if metrics.Metrics.LeadTime == 0 {
			metrics.Metrics.LeadTime = 1
		}

		if metrics.Metrics.CycleTime == 0 {
			metrics.Metrics.CycleTime = 1
		}
		// Calculate Flow Efficiency
		metrics.Metrics.FlowEfficiency = (float64(metrics.Metrics.CycleTime) - float64(metrics.Metrics.BlockedTime)) * 100 / float64(metrics.Metrics.CycleTime)
		metricsPerTask = append(metricsPerTask, metrics)
	}

	return metricsPerTask
}
