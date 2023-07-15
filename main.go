package main

import (
	"fmt"
	"log"

	"github.com/lucasvillalbaar/clickup-metrics/pkg/configuration"
	"github.com/lucasvillalbaar/clickup-metrics/pkg/data"
	"github.com/lucasvillalbaar/clickup-metrics/pkg/data/clickup"
	"github.com/lucasvillalbaar/clickup-metrics/pkg/metrics"
)

func contains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

func setStatusesForLeadTimeCalculation(wfStatuses *map[string]metrics.Status) {
	//All status are valid for Lead Time calculation
	for key := range *wfStatuses {
		status := (*wfStatuses)[key]
		status.IsLeadTimeCalculable = true
		(*wfStatuses)[key] = status
	}
}

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

func main() {
	err := configuration.LoadEnvironmentVariables()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}
	cli := clickup.Init(configuration.GetEnvironmentVariables().ApiKey, configuration.GetEnvironmentVariables().Token)

	data.SetDataSource(cli)

	wf := createWorkflow(cli)

	taskInfo := data.GetTaskByID("85zt8cyjd")
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

	for _, result := range metricsPerTask {
		fmt.Printf("Task ID: %s | %s | %s | %s\n", result.TaskInfo.Id, result.TaskInfo.Name, result.TaskInfo.StartDate, result.TaskInfo.DueDate)
		fmt.Printf("Lead Time: %d | Cycle Time: %d | Blocked Time: %d | Flow Efficiency: %.2f\n", result.Metrics.LeadTime, result.Metrics.CycleTime, result.Metrics.BlockedTime, result.Metrics.FlowEfficiency)
	}

}
