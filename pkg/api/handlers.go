package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/gorilla/mux"
)

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
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	// Estructura para almacenar los datos
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
			"templates/footer.gohtml")

	//tmpl, err := template.New("dashboard").Parse(htmlTemplate)
	if err != nil {
		log.Println("Error when reading template: ", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	// Datos de ejemplo
	datosList := []Datos{
		{
			ID:             "85zruhhbb",
			CustomID:       "CORE-2",
			Nombre:         "Instalar app mobile en el celular",
			FechaInicio:    "2023-04-05",
			FechaFin:       "2023-04-24",
			LeadTime:       92,
			CycleTime:      29,
			BlockedTime:    63,
			FlowEfficiency: "-117,00%",
		},
		{
			ID:             "85zruhhbb",
			CustomID:       "CORE-234",
			Nombre:         "Inssdsdlular",
			FechaInicio:    "2023-04-05",
			FechaFin:       "2023-04-24",
			LeadTime:       92,
			CycleTime:      29,
			BlockedTime:    63,
			FlowEfficiency: "-117,00%",
		},
		{
			ID:             "85zruhhbb",
			CustomID:       "CORE-234",
			Nombre:         "Inssdsdlular",
			FechaInicio:    "2023-04-05",
			FechaFin:       "2023-04-24",
			LeadTime:       92,
			CycleTime:      29,
			BlockedTime:    63,
			FlowEfficiency: "-117,00%",
		},
	}

	type ChartData struct {
		ChartID    string
		ChartLabel string
		Data       []int
		Labels     []string
	}

	leadTimeData := ChartData{
		ChartID:    "lead-time-chart",
		ChartLabel: "Lead Time",
		Data:       []int{15, 25, 20},
		Labels:     []string{"CORE-2", "CORE-234", "CORE-233"},
	}

	cycleTimeData := ChartData{
		ChartID:    "cycle-time-chart",
		ChartLabel: "Cycle Time",
		Data:       []int{10, 25, 15},
		Labels:     []string{"CORE-2", "CORE-234", "CORE-233"},
	}

	blockedTimeData := ChartData{
		ChartID:    "blocked-time-chart",
		ChartLabel: "Blocked Time",
		Data:       []int{1, 5, 0},
		Labels:     []string{"CORE-2", "CORE-234", "CORE-233"},
	}

	flowEfficiencyData := ChartData{
		ChartID:    "flow-efficiency-chart",
		ChartLabel: "Flow Efficiency",
		Data:       []int{100, 80, 83},
		Labels:     []string{"CORE-2", "CORE-234", "CORE-233"},
	}

	mergeRequestTimeToMerge := ChartData{
		ChartID:    "merge-request-time-to-merge-chart",
		ChartLabel: "Merge Request - Time To Merge",
		Data:       []int{39, 4, 0, 1, 1, 0, 1, 2},
		Labels:     []string{"<1d", "1d", "2d", "3d", "4d", "5d", "6d", "+7d"},
	}

	mergeRequestSize := ChartData{
		ChartID:    "merge-request-size-chart",
		ChartLabel: "Merge Request - Size",
		Data:       []int{20, 10, 11, 4},
		Labels:     []string{"Small (50)", "Medium (51-200)", "Large (201-500)", "Very Large (+500)"},
	}

	data := struct {
		DatosList               []Datos
		StartDate               string
		EndDate                 string
		LeadTimeData            ChartData
		CycleTimeData           ChartData
		BlockedTimeData         ChartData
		FlowEfficiencyData      ChartData
		MergeRequestTimeToMerge ChartData
		MergeRequestSize        ChartData
	}{
		DatosList:               datosList,
		StartDate:               startDate,
		EndDate:                 endDate,
		LeadTimeData:            leadTimeData,
		CycleTimeData:           cycleTimeData,
		BlockedTimeData:         blockedTimeData,
		FlowEfficiencyData:      flowEfficiencyData,
		MergeRequestTimeToMerge: mergeRequestTimeToMerge,
		MergeRequestSize:        mergeRequestSize,
	}

	// Rellenar la plantilla con los datos y escribir la respuesta HTTP
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println("Error when executing template: ", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
}

func toJson(v interface{}) (string, error) {
	jsonData, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
