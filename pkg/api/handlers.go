package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/lucasvillalbaar/clickup-metrics/pkg/mergerequests"
)

type ChartData struct {
	ChartID    string
	ChartLabel string
	Data       []int
	Labels     []string
}
type DashboardData struct {
	StartDate               string
	EndDate                 string
	Tickets                 string
	AvgLeadTime             int
	AvgCycleTime            int
	AvgBlockedTime          int
	AvgFlowEfficiency       float64
	IsClickupTokenExpired   bool
	TaskMetrics             []TaskMetricsResponse
	LeadTimeData            ChartData
	CycleTimeData           ChartData
	BlockedTimeData         ChartData
	FlowEfficiencyData      ChartData
	MergeRequests           []mergerequests.MergeRequest
	MergeRequestTimeToMerge ChartData
	MergeRequestSize        ChartData
}

type Datos struct {
	ID             string
	CustomID       string
	Nombre         string
	FechaInicio    string
	FechaFin       string
	LeadTime       int
	CycleTime      int
	BlockedTime    int
	FlowEfficiency string
}

// getTaskMetricsHandler is the handler function for the GET /metrics/{task_id} endpoint.
// It retrieves the task metrics for the specified task ID and returns them as JSON.
func getTaskMetricsHandler(w http.ResponseWriter, r *http.Request) {
	// Set the response headers
	w.Header().Set("Content-Type", "application/json")
	// Get the task ID from the request parameters
	vars := mux.Vars(r)
	taskID := vars["task_id"]

	// Retrieve the task metrics for the specified task ID
	taskMetrics, err := getTaskMetrics(r.Context(), taskID)
	if err != nil {
		switch err.Error() {
		case "token is expired":
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"error": "Clickup Token expired. Get a new one and do the request again"}`)
			return
		case "api key is expired or is not valid":
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"error": "Api Key is expired or is not valid. Get a new one and do the request again"}`)
			return
		}
	}

	// Marshal the task metrics to JSON
	jsonData, err := json.Marshal(taskMetrics)
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set the response headers
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON data to the response
	_, err = w.Write(jsonData)
	if err != nil {
		log.Println("Error writing response:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func getDashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Extract query parameters from the request
	startDateParam := r.URL.Query().Get("start_date")
	endDateParam := r.URL.Query().Get("end_date")
	ticketsParam := r.URL.Query().Get("tickets")

	tickets, err := url.QueryUnescape(ticketsParam)
	if err != nil {
		// Maneja el error si la decodificación falla
		http.Error(w, "Error when docoding tickets param'", http.StatusBadRequest)
		return
	}

	// Register the toJson function as a custom template function
	funcMap := template.FuncMap{
		"toJson": toJson,
	}

	// Compilar la plantilla desde el archivo
	tmpl, err := template.New("dashboard.gohtml").Funcs(funcMap).
		ParseFiles("templates/dashboard.gohtml",
			"templates/average_metrics.gohtml",
			"templates/line_chart.gohtml",
			"templates/bar_chart.gohtml",
			"templates/no_data.gohtml",
			"templates/tickets_table.gohtml",
			"templates/merge_requests_table.gohtml",
			"templates/scripts.gohtml",
			"templates/footer.gohtml")

	if err != nil {
		log.Println("Error when reading template: ", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	data, err := getDashboardData(startDateParam, endDateParam, tickets)

	// Rellenar la plantilla con los datos y escribir la respuesta HTTP
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println("Error when executing template: ", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
}

func getClickUpData(result *DashboardData, tickets string) {
	if tickets == "" {
		return
	}

	token := os.Getenv("CLICKUP_TOKEN")
	if token == "" {
		result.IsClickupTokenExpired = true
	}
	ctx := context.WithValue(context.TODO(), ContextClickUpToken, "Bearer "+token)

	ticketsSlice := strings.Split(tickets, ",")
	leadTimeDataSlice := []int{}
	leadTimeLabelsSlice := []string{}
	cycleTimeDataSlice := []int{}
	cycleTimeLabelsSlice := []string{}
	blockedTimeDataSlice := []int{}
	blockedTimeLabelsSlice := []string{}
	flowEfficiencyDataSlice := []int{}
	flowEfficiencyLabalsSlice := []string{}

	for _, ticketId := range ticketsSlice {
		if result.IsClickupTokenExpired == true {
			continue
		}
		ticketIdStr := strings.ReplaceAll(ticketId, "#", "")
		ticketIdStr = strings.ReplaceAll(ticketIdStr, " ", "")
		ticketMetrics, err := getTaskMetrics(ctx, ticketIdStr)
		if err != nil {
			if err.Error() == "api key is expired or is not valid" || err.Error() == "token is expired" {
				result.IsClickupTokenExpired = true
			}
			log.Println(err)
			continue
		}
		result.AvgLeadTime = result.AvgLeadTime + ticketMetrics.LeadTime
		result.AvgCycleTime = result.AvgCycleTime + ticketMetrics.CycleTime
		result.AvgBlockedTime = result.AvgBlockedTime + ticketMetrics.BlockedTime
		result.AvgFlowEfficiency = result.AvgFlowEfficiency + ticketMetrics.FlowEfficiency
		leadTimeDataSlice = append(leadTimeDataSlice, ticketMetrics.LeadTime)
		leadTimeLabelsSlice = append(leadTimeLabelsSlice, ticketMetrics.CustomId)
		cycleTimeDataSlice = append(cycleTimeDataSlice, ticketMetrics.CycleTime)
		cycleTimeLabelsSlice = append(cycleTimeLabelsSlice, ticketMetrics.CustomId)
		blockedTimeDataSlice = append(blockedTimeDataSlice, ticketMetrics.BlockedTime)
		blockedTimeLabelsSlice = append(blockedTimeLabelsSlice, ticketMetrics.CustomId)
		flowEfficiencyDataSlice = append(flowEfficiencyDataSlice, int(ticketMetrics.FlowEfficiency))
		flowEfficiencyLabalsSlice = append(flowEfficiencyLabalsSlice, ticketMetrics.CustomId)
		result.TaskMetrics = append(result.TaskMetrics, ticketMetrics)
	}

	result.AvgLeadTime = result.AvgLeadTime / len(ticketsSlice)
	result.AvgCycleTime = result.AvgCycleTime / len(ticketsSlice)
	result.AvgBlockedTime = result.AvgBlockedTime / len(ticketsSlice)
	result.AvgFlowEfficiency = result.AvgFlowEfficiency / float64(len(ticketsSlice))

	result.LeadTimeData = ChartData{
		ChartID:    "lead-time-chart",
		ChartLabel: "Lead Time",
		Data:       leadTimeDataSlice,
		Labels:     leadTimeLabelsSlice,
	}

	result.CycleTimeData = ChartData{
		ChartID:    "cycle-time-chart",
		ChartLabel: "Cycle Time",
		Data:       cycleTimeDataSlice,
		Labels:     cycleTimeLabelsSlice,
	}

	result.BlockedTimeData = ChartData{
		ChartID:    "blocked-time-chart",
		ChartLabel: "Blocked Time",
		Data:       blockedTimeDataSlice,
		Labels:     blockedTimeLabelsSlice,
	}

	result.FlowEfficiencyData = ChartData{
		ChartID:    "flow-efficiency-chart",
		ChartLabel: "Flow Efficiency",
		Data:       flowEfficiencyDataSlice,
		Labels:     flowEfficiencyLabalsSlice,
	}
}

func getGitLabData(result *DashboardData, startDate string, endDate string) {
	if startDate == "" || endDate == "" {
		return
	}
	gitlabToken := os.Getenv("GITLAB_TOKEN")
	mrsCli := mergerequests.NewGitlabClient("5908940", "CORE", gitlabToken)
	mrsSlice, _ := mrsCli.GetMergeRequestsMergedBetween(startDate, endDate)

	result.MergeRequests = mrsSlice

	result.MergeRequestTimeToMerge = initMergeRequestTimeToMergeChartData()
	result.MergeRequestSize = initMergeRequestSizeChartData()
	for _, mr := range result.MergeRequests {
		result.MergeRequestTimeToMerge = appendMergeRequestTimeToMergeChartData(result.MergeRequestTimeToMerge, mr)
		result.MergeRequestSize = appendMergeRequestSizeChartData(result.MergeRequestSize, mr)
	}
}

func getDashboardData(startDate string, endDate string, tickets string) (*DashboardData, error) {
	result := &DashboardData{
		StartDate:             startDate,
		EndDate:               endDate,
		Tickets:               tickets,
		IsClickupTokenExpired: false,
	}

	getClickUpData(result, tickets)

	getGitLabData(result, startDate, endDate)

	return result, nil
}

func initMergeRequestSizeChartData() ChartData {
	return ChartData{
		ChartID:    "merge-request-size-chart",
		ChartLabel: "Merge Request - Size",
		Data:       []int{0, 0, 0, 0},
		Labels:     []string{"Small (50)", "Medium (51-200)", "Large (201-500)", "Very Large (+500)"},
	}
}

func initMergeRequestTimeToMergeChartData() ChartData {
	return ChartData{
		ChartID:    "merge-request-time-to-merge-chart",
		ChartLabel: "Merge Request - Time To Merge",
		Data:       []int{0, 0, 0, 0, 0, 0, 0, 0},
		Labels:     []string{"<1d", "1d", "2d", "3d", "4d", "5d", "6d", "+7d"},
	}
}

func appendMergeRequestSizeChartData(size ChartData, mr mergerequests.MergeRequest) ChartData {
	var index int

	switch {
	case mr.Size <= 50:
		index = 0
	case mr.Size >= 51 && mr.Size <= 200:
		index = 1
	case mr.Size >= 201 && mr.Size <= 500:
		index = 2
	default:
		index = 3
	}

	size.Data[index]++

	return size
}

func appendMergeRequestTimeToMergeChartData(timeToMerge ChartData, mr mergerequests.MergeRequest) ChartData {
	const MoreThan7Days = 7
	days := mr.TimeToMerge

	switch {
	case days < 0:
		return timeToMerge
	case days >= 7:
		timeToMerge.Data[MoreThan7Days]++
	default:
		timeToMerge.Data[days]++
	}

	return timeToMerge
}

func toJson(v interface{}) (string, error) {
	jsonData, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

type TokenRequestBody struct {
	Token string `json:"token"`
}

type TokenResponse struct {
	Message string `json:"message"`
}

func setTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading the request body", http.StatusBadRequest)
		return
	}

	// Decode the JSON body into a TokenRequestBody struct
	var requestBody TokenRequestBody
	if err := json.Unmarshal(body, &requestBody); err != nil {
		http.Error(w, "Error decoding the JSON body", http.StatusBadRequest)
		return
	}

	// Get the token from the struct
	tokenValue := requestBody.Token

	// Set the token as an environment variable
	os.Setenv("CLICKUP_TOKEN", tokenValue)

	// Respond with a JSON success message
	responseMessage := TokenResponse{Message: "Token saved successfully"}
	responseJSON, err := json.Marshal(responseMessage)
	if err != nil {
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
		return
	}

	// Respond with a success message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}
