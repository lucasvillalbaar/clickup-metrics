package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"text/template"

	"github.com/gorilla/mux"
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
	AvgLeadTime             int
	AvgCycleTime            int
	AvgBlockedTime          int
	AvgFlowEfficiency       float64
	TaskMetrics             []TaskMetricsResponse
	LeadTimeData            ChartData
	CycleTimeData           ChartData
	BlockedTimeData         ChartData
	FlowEfficiencyData      ChartData
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
			"templates/tickets_table.gohtml",
			"templates/scripts.gohtml",
			"templates/footer.gohtml")

	if err != nil {
		log.Println("Error when reading template: ", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	data := getDashboardData(startDateParam, endDateParam, tickets)

	// Rellenar la plantilla con los datos y escribir la respuesta HTTP
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println("Error when executing template: ", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
}

func getDashboardData(startDate string, endDate string, tickets string) *DashboardData {
	if startDate == "" || endDate == "" || tickets == "" {
		return &DashboardData{}
	}

	result := &DashboardData{
		StartDate: startDate,
		EndDate:   endDate,
	}

	ctx := context.WithValue(context.TODO(), ContextClickUpToken, "Bearer "+"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6IjAifQ.eyJ1c2VyIjo2MTQ4MjM5NCwidmFsaWRhdGVkIjp0cnVlLCJ3c19rZXkiOjU1NzgyMDQyMTcsInNlc3Npb25fdG9rZW4iOnRydWUsImlhdCI6MTY5NTk5NjczNiwiZXhwIjoxNjk2MTY5NTM2fQ.HgHw2WdjENjxdwgbE3TnzqiacsK6lIo7rWfrceV7iac")

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
		ticketIdStr := strings.ReplaceAll(ticketId, "#", "")
		ticketIdStr = strings.ReplaceAll(ticketIdStr, " ", "")
		ticketMetrics, err := getTaskMetrics(ctx, ticketIdStr)
		if err != nil {
			log.Println("Error in getDashboardData: ", err)
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
		blockedTimeDataSlice = append(blockedTimeDataSlice, ticketMetrics.LeadTime)
		blockedTimeLabelsSlice = append(blockedTimeLabelsSlice, ticketMetrics.CustomId)
		flowEfficiencyDataSlice = append(flowEfficiencyDataSlice, ticketMetrics.LeadTime)
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

	result.MergeRequestTimeToMerge = ChartData{
		ChartID:    "merge-request-time-to-merge-chart",
		ChartLabel: "Merge Request - Time To Merge",
		Data:       []int{39, 4, 0, 1, 1, 0, 1, 2},
		Labels:     []string{"<1d", "1d", "2d", "3d", "4d", "5d", "6d", "+7d"},
	}

	result.MergeRequestSize = ChartData{
		ChartID:    "merge-request-size-chart",
		ChartLabel: "Merge Request - Size",
		Data:       []int{20, 10, 11, 4},
		Labels:     []string{"Small (50)", "Medium (51-200)", "Large (201-500)", "Very Large (+500)"},
	}

	return result
}

func toJson(v interface{}) (string, error) {
	jsonData, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
