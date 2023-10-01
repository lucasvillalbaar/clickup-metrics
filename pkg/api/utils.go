package api

import (
	"strconv"
	"time"

	"github.com/lucasvillalbaar/clickup-metrics/pkg/data"
	"github.com/lucasvillalbaar/clickup-metrics/pkg/data/clickup"
	"github.com/lucasvillalbaar/clickup-metrics/pkg/metrics"
)

// contains checks if a value is present in a list.
func contains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

// setStatusesForLeadTimeCalculation sets all statuses as calculable for Lead Time.
func setStatusesForLeadTimeCalculation(wfStatuses *map[string]metrics.Status) {
	// All statuses are valid for Lead Time calculation
	for key := range *wfStatuses {
		status := (*wfStatuses)[key]
		status.IsLeadTimeCalculable = true
		(*wfStatuses)[key] = status
	}
}

// setStatusesForCycleTimeCalculation sets the specified statuses as calculable for Cycle Time.
func setStatusesForCycleTimeCalculation(wfStatuses *map[string]metrics.Status, validStatuses ...string) {
	for key := range *wfStatuses {
		status := (*wfStatuses)[key]
		if contains(validStatuses, status.Name) {
			status.IsCycleTimeCalculable = true
		} else {
			status.IsCycleTimeCalculable = false
		}
		(*wfStatuses)[key] = status
	}
}

// createWorkflow creates a metrics.Workflow based on the data from the data.Data object.
func createWorkflow(d data.Data) metrics.Workflow {
	wf := metrics.Workflow{
		Statuses: make(map[string]metrics.Status),
	}

	for _, status := range d.GetWorkflow().Statuses {
		mts := metrics.Status{
			Name:       status.Name,
			Pending:    status.Pending,
			InProgress: status.InProgress,
			Blocked:    status.Blocked,
			Done:       status.Done,
		}
		wf.Statuses[status.Name] = mts
	}

	setStatusesForLeadTimeCalculation(&wf.Statuses)
	setStatusesForCycleTimeCalculation(&wf.Statuses,
		clickup.StatusInDevelopment,
		clickup.StatusReadyToDeploy,
		clickup.StatusDeployed,
		clickup.StatusComplete)

	return wf
}

// prepareHistory prepares the history transitions for a task.
func prepareHistory(taskInfo data.TaskInfo) []metrics.Transition {
	history := []metrics.Transition{}

	for _, transition := range taskInfo.History {
		history = append(history, metrics.Transition{
			Before: transition.Before.Status,
			After:  transition.After.Status,
			Date:   transition.Date,
		})
	}

	return history
}

// ConvertUnixMillisToString converts a Unix timestamp in milliseconds (string format) to a formatted date string.
func ConvertUnixMillisToString(unixMillis string) (string, error) {
	// Convert the string to an int64
	unixMillisInt, err := strconv.ParseInt(unixMillis, 10, 64)
	if err != nil {
		return "", err
	}

	// Create a time.Time object using the milliseconds value
	t := time.Unix(0, unixMillisInt*int64(time.Millisecond))

	// Format the date and time using the desired format "YYYY-MM-DD"
	formatted := t.Format("2006-01-02")

	return formatted, nil
}
