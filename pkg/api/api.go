package api

import (
	"log"

	"github.com/gorilla/mux"
	"github.com/lucasvillalbaar/clickup-metrics/pkg/configuration"
	"github.com/lucasvillalbaar/clickup-metrics/pkg/data"
	"github.com/lucasvillalbaar/clickup-metrics/pkg/data/clickup"
	"github.com/lucasvillalbaar/clickup-metrics/pkg/metrics"
)

type TaskMetricsResponse struct {
	Id             string  `json:"id"`
	CustomId       string  `json:"custom_id"`
	Name           string  `json:"name"`
	StartDate      string  `json:"start_date"`
	DueDate        string  `json:"due_date"`
	LeadTime       int     `json:"lead_time"`
	CycleTime      int     `json:"cycle_time"`
	BlockedTime    int     `json:"blocked_time"`
	FlowEfficiency float64 `json:"flow_efficiency"`
}

var wf metrics.Workflow

// initialization loads environment variables, initializes the data source, and creates the workflow
func initialization() {
	err := configuration.LoadEnvironmentVariables()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}
	cli := clickup.Init(configuration.GetEnvironmentVariables().ApiKey, configuration.GetEnvironmentVariables().Token)

	data.SetDataSource(cli)

	wf = createWorkflow(cli)
}

// Init initializes the API router and sets up the routes
func Init() *mux.Router {
	initialization()
	router := mux.NewRouter()

	router.HandleFunc("/metrics/{task_id}", getTaskMetricsHandler).Methods("GET")

	return router
}

// getTaskMetrics retrieves the metrics for a specific task
func getTaskMetrics(taskID string) TaskMetricsResponse {
	taskInfo := data.GetTaskByID(taskID)
	history := prepareHistory(*taskInfo)

	tasks := []metrics.TaskInfo{}
	tasks = append(tasks, metrics.TaskInfo{
		Id:        taskInfo.Id,
		Name:      taskInfo.Name,
		StartDate: taskInfo.StartDate,
		DueDate:   taskInfo.DueDate,
		History:   history,
	})

	metrics.InitMetrics(wf)
	metrics.SetInfo(tasks)

	metricsPerTask := metrics.CalculateMetrics()
	if len(metricsPerTask) != 0 {
		result := metricsPerTask[0]
		startDate, _ := ConvertUnixMillisToString(result.TaskInfo.StartDate)
		dueDate, _ := ConvertUnixMillisToString(result.TaskInfo.DueDate)
		return TaskMetricsResponse{
			Id:             result.TaskInfo.Id,
			CustomId:       taskInfo.CustomId,
			Name:           result.TaskInfo.Name,
			StartDate:      startDate,
			DueDate:        dueDate,
			LeadTime:       result.Metrics.LeadTime,
			CycleTime:      result.Metrics.CycleTime,
			BlockedTime:    result.Metrics.BlockedTime,
			FlowEfficiency: result.Metrics.FlowEfficiency,
		}
	}

	return TaskMetricsResponse{}
}
