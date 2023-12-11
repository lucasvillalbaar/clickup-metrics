package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lucasvillalbaar/clickup-metrics/pkg/configuration"
	"github.com/lucasvillalbaar/clickup-metrics/pkg/data"
	"github.com/lucasvillalbaar/clickup-metrics/pkg/data/clickup"
	"github.com/lucasvillalbaar/clickup-metrics/pkg/metrics"
)

type TimeSpentPerStateResponse struct {
	Name      string `json:"name"`
	TimeSpent int    `json:"time_spent"`
}
type TaskMetricsResponse struct {
	Id             string         `json:"id"`
	CustomId       string         `json:"custom_id"`
	Name           string         `json:"name"`
	StartDate      string         `json:"start_date"`
	DueDate        string         `json:"due_date"`
	LeadTime       int            `json:"lead_time"`
	CycleTime      int            `json:"cycle_time"`
	BlockedTime    int            `json:"blocked_time"`
	FlowEfficiency float64        `json:"flow_efficiency"`
	Statuses       []data.History `json:"statuses"`
}

const (
	ContextClickUpToken = "clickup_token"
)

var wf metrics.Workflow

// initialization loads environment variables, initializes the data source, and creates the workflow
func configureDataSource() {
	err := configuration.LoadEnvironmentVariables()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	cli := clickup.Init(configuration.GetEnvironmentVariables().ApiKey)

	data.SetDataSource(cli)

	wf = createWorkflow(cli)
}

// Init initializes the API router and sets up the routes
func Init() *mux.Router {
	router := mux.NewRouter()

	router.Use(authInterceptor)
	router.HandleFunc("/healthcheck", getHealthCheck).Methods("GET")
	router.HandleFunc("/dashboard", getDashboardHandler).Methods("GET")

	router.HandleFunc("/metrics/{task_id}", getTaskMetricsHandler).Methods("GET")

	// Ruta para servir archivos estáticos (por ejemplo, CSS)
	staticFileServer := http.FileServer(http.Dir("./static"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", staticFileServer))

	return router
}

func getHealthCheck(w http.ResponseWriter, r *http.Request) {
	// Set content type to JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 status

	response := map[string]string{
		"status":  "OK",
		"message": "API is running",
	}

	// Encode and send JSON response
	json.NewEncoder(w).Encode(response)
}

// authInterceptor es el interceptor que se ejecutará antes de manejar la solicitud
func authInterceptor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: add api key

		next.ServeHTTP(w, r)
	})
}

// getTaskMetrics retrieves the metrics for a specific task
func getTaskMetrics(taskID string) (TaskMetricsResponse, error) {
	configureDataSource()
	taskInfo, err := data.GetTaskByID(taskID)
	if err != nil {
		log.Println(err)
		return TaskMetricsResponse{}, err
	}

	tasks := []metrics.TaskInfo{}
	tasks = append(tasks, metrics.TaskInfo{
		Id:        taskInfo.Id,
		Name:      taskInfo.Name,
		StartDate: taskInfo.StartDate,
		DueDate:   taskInfo.DueDate,
		History:   taskInfo.History,
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
			Statuses:       result.TaskInfo.History,
		}, nil
	}

	return TaskMetricsResponse{}, nil
}
