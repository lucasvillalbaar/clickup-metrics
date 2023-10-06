package mergerequests

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	GitlabURLGetMergeRequests       = "https://gitlab.com/api/v4/groups/%s/merge_requests?scope=all&created_after=%sT00:00:00.000Z&created_before=%sT23:59:59.999Z&search=%s&in=title&state=merged"
	GitlabURLGetMergeRequestChanges = "https://gitlab.com/api/v4/projects/%d/merge_requests/%d/changes"
)

type Change struct {
	Diff string `json:"diff"`
}
type MergeRequestChange struct {
	Changes []Change `json:"changes"`
}
type MergeRequest struct {
	ID          int    `json:"id"`
	IID         int    `json:"iid"`
	ProjectID   int    `json:"project_id"`
	Title       string `json:"title"`
	CreatedAt   string `json:"created_at"`
	MergedAt    string `json:"merged_at"`
	TimeToMerge int    `json:"time_to_merge"`
	Size        int    `json:"size"`
}

type GitlabClient struct {
	GroupID string
	Team    string
	Token   string
}

func NewGitlabClient(groupID, team, token string) *GitlabClient {
	return &GitlabClient{
		GroupID: groupID,
		Team:    team,
		Token:   token,
	}
}

func (c *GitlabClient) GetMergeRequestsMergedBetween(startDate string, endDate string) ([]MergeRequest, error) {
	url := fmt.Sprintf(GitlabURLGetMergeRequests, c.GroupID, startDate, endDate, c.Team)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", c.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New("Resource not found")
	}

	var mergeRequests []MergeRequest
	if err := json.NewDecoder(resp.Body).Decode(&mergeRequests); err != nil {
		log.Println(err)
		return nil, err
	}

	var result []MergeRequest
	var wg sync.WaitGroup

	for _, mr := range mergeRequests {
		mrCopy := mr
		wg.Add(1)
		go func(mr MergeRequest) {
			defer wg.Done()

			change, _ := c.GetMergeRequestChanges(mr.ProjectID, mr.IID)
			result = append(result, MergeRequest{
				ID:          mr.ID,
				IID:         mr.IID,
				ProjectID:   mr.ProjectID,
				Title:       mr.Title,
				Size:        calculateNetChangesSize(change),
				TimeToMerge: getTimeToMerge(&mr),
				CreatedAt:   formatDate(mr.CreatedAt),
				MergedAt:    formatDate(mr.MergedAt),
			})
		}(mrCopy)
	}
	wg.Wait()
	sortByCreatedAt(result)
	return result, nil
}

// sortByCreatedAt sorts a slice of MergeRequest structs by their CreatedAt dates in ascending order.
// It takes a slice of MergeRequest structs as input and sorts it in-place.
func sortByCreatedAt(mrs []MergeRequest) {
	sort.Slice(mrs, func(i, j int) bool {
		return mrs[j].CreatedAt > mrs[i].CreatedAt
	})
}

func (c *GitlabClient) GetMergeRequestChanges(projectID int, iid int) (MergeRequestChange, error) {
	url := fmt.Sprintf(GitlabURLGetMergeRequestChanges, projectID, iid)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return MergeRequestChange{}, err
	}
	req.Header.Set("PRIVATE-TOKEN", c.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return MergeRequestChange{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		log.Println("Resource not found")
		return MergeRequestChange{}, errors.New("Resource not found")
	}

	var changes MergeRequestChange

	if err := json.NewDecoder(resp.Body).Decode(&changes); err != nil {
		log.Println(err)
		return MergeRequestChange{}, err
	}

	return changes, nil
}

func calculateNetChangesSize(change MergeRequestChange) int {
	netAdditions := 0
	netDeletions := 0
	for _, change := range change.Changes {
		lines := strings.Split(change.Diff, "\n")

		for _, line := range lines {
			if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
				netAdditions++
			} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
				netDeletions++
			}
		}
	}

	netChangesSize := netAdditions + netDeletions
	return netChangesSize
}

func getTimeToMerge(mr *MergeRequest) int {
	createdAt, _ := time.Parse("2006-01-02T15:04:05.999999Z07:00", mr.CreatedAt)
	mergedAt, _ := time.Parse("2006-01-02T15:04:05.999999Z07:00", mr.MergedAt)

	timeDifference := mergedAt.Sub(createdAt)
	daysDifference := int(timeDifference.Hours() / 24)

	weekends := (daysDifference + int(createdAt.Weekday()) + 1) / 7 * 2

	if createdAt.Weekday() == time.Sunday {
		weekends--
	}
	if mergedAt.Weekday() == time.Saturday {
		weekends--
	}

	totalDaysWithoutWeekends := daysDifference - weekends

	return totalDaysWithoutWeekends
}

func (mr *MergeRequest) GetSize() int {
	return mr.Size
}

func formatDate(dateStr string) string {
	// Parse the date string
	parsedTime, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return "Invalid Date"
	}

	// Format the date as "YYYY-MM-DD HH:MM"
	formattedDate := parsedTime.Format("2006-01-02 15:04")

	return formattedDate
}
