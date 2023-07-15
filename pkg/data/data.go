package data

import "time"

type TransitionState struct {
	Status string
}

// Structures to parse JSON
type Transition struct {
	Before TransitionState
	After  TransitionState
	Date   string
}

type TaskHeaderData struct {
	Id        string
	Name      string
	CustomId  string
	StartDate string
	DueDate   string
}

type TaskInfo struct {
	TaskHeaderData
	History []Transition
}

type Filter struct {
	ProjectID       string
	TaskType        []string
	DueDateAfter    time.Time
	OnlyClosedTasks bool
}

type Session struct {
	ServiceName string
}

type Status struct {
	Name       string
	Pending    bool
	InProgress bool
	Blocked    bool
	Done       bool
}
type Workflow struct {
	Statuses []Status
}

type Data interface {
	GetTasksWithFilter(filter Filter) []TaskHeaderData
	GetTaskByID(id string) *TaskInfo
	GetHistoryPerTask(ids []string) *map[string]TaskInfo
	GetWorkflow() *Workflow
}

var data Data

func SetDataSource(d Data) {
	data = d
}

func GetTasksWithFilter(filter Filter) []TaskHeaderData {
	return data.GetTasksWithFilter(filter)
}

func GetTaskByID(id string) *TaskInfo {
	return data.GetTaskByID(id)
}

func GetHistoryPerTask(ids []string) *map[string]TaskInfo {
	return data.GetHistoryPerTask(ids)
}
