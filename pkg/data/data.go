package data

import "time"

type History struct {
	Status string
	Time   int
	Since  string
}

type TotalTime struct {
	ByMinute int    `json:"by_minute"`
	Since    string `json:"since"`
}

type StatusHistory struct {
	Status    string    `json:"status"`
	TotalTime TotalTime `json:"total_time"`
}

type TimeInStatusResponse struct {
	StatusHistory []StatusHistory `json:"status_history"`
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
	History []History
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
	GetTaskByID(id string) (*TaskInfo, error)
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

func GetTaskByID(id string) (*TaskInfo, error) {
	return data.GetTaskByID(id)
}

func GetHistoryPerTask(ids []string) *map[string]TaskInfo {
	return data.GetHistoryPerTask(ids)
}
