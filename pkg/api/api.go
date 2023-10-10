package api

import (
	"context"
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
	Id             string                      `json:"id"`
	CustomId       string                      `json:"custom_id"`
	Name           string                      `json:"name"`
	StartDate      string                      `json:"start_date"`
	DueDate        string                      `json:"due_date"`
	LeadTime       int                         `json:"lead_time"`
	CycleTime      int                         `json:"cycle_time"`
	BlockedTime    int                         `json:"blocked_time"`
	FlowEfficiency float64                     `json:"flow_efficiency"`
	Statuses       []TimeSpentPerStateResponse `json:"statuses"`
}

const (
	ContextClickUpToken = "clickup_token"
)

var wf metrics.Workflow

// initialization loads environment variables, initializes the data source, and creates the workflow
func configureDataSource(ctx context.Context) {
	err := configuration.LoadEnvironmentVariables()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}
	token := ctx.Value(ContextClickUpToken).(string)
	if token == "" {
		log.Println("Clickup token has not been configured")
	}
	cli := clickup.Init(configuration.GetEnvironmentVariables().ApiKey, token)

	data.SetDataSource(cli)

	wf = createWorkflow(cli)
}

// Init initializes the API router and sets up the routes
func Init() *mux.Router {
	router := mux.NewRouter()

	router.Use(authInterceptor)
	router.HandleFunc("/healthcheck", getHealthCheck).Methods("GET")
	router.HandleFunc("/dashboard", getDashboardHandler).Methods("GET")
	router.HandleFunc("/token", setTokenHandler).Methods("POST")

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
		token := r.Header.Get("Authorization")

		ctx := context.WithValue(r.Context(), ContextClickUpToken, token)

		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// getTaskMetrics retrieves the metrics for a specific task
func getTaskMetrics(ctx context.Context, taskID string) (TaskMetricsResponse, error) {
	configureDataSource(ctx)
	taskInfo, err := data.GetTaskByID(taskID)
	if err != nil {
		return TaskMetricsResponse{}, err
	}

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
		statuses := getTimeSpentPerState(&result.MetricsPerState)
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
			Statuses:       statuses,
		}, nil
	}

	return TaskMetricsResponse{}, nil
}

// getTimeSpentPerState calculates the time spent per state and returns a slice of TimeSpentPerStateResponse.
// It takes a pointer to a MetricsPerState map as input.
// Each entry in the map represents a state and its corresponding time spent.
// The function iterates through the map, extracts the state name and time spent,
// and creates TimeSpentPerStateResponse objects, which are then appended to a result slice.
// Finally, the function returns the result slice containing the calculated data.
//
// Parameters:
//   - ms (*metrics.MetricsPerState): A pointer to a MetricsPerState map.
//
// Returns:
//
//	([]TimeSpentPerStateResponse): A slice of TimeSpentPerStateResponse objects
//	                               representing the time spent per state.
func getTimeSpentPerState(ms *metrics.MetricsPerState) []TimeSpentPerStateResponse {
	result := []TimeSpentPerStateResponse{}
	for key, entry := range *ms {
		result = append(result, TimeSpentPerStateResponse{
			Name:      key,
			TimeSpent: entry.TimeSpent,
		})
	}
	return result
}
