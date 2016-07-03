package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	mp "github.com/mackerelio/go-mackerel-plugin-helper"
)

// RedashPlugin mackerel plugin
type RedashPlugin struct {
	URL     string
	APIKey  string
	Prefix  string
	Timeout uint
}

// RedashStatus struct for RedashPlugin mackerel plugin
type RedashStatus struct {
	DashboardsCount         uint64        `json:"dashboards_count"`
	Manager                 RedashManager `json:"manager"`
	QueriesCount            uint64        `json:"queries_count"`
	QueryResultsCount       uint64        `json:"query_results_count"`
	RedisUsedMemory         string        `json:"redis_used_memory"`
	UnusedQueryResultsCount uint64        `json:"unused_query_results_count"`
	Version                 string        `json:"version"`
	WidgetsCount            uint64        `json:"widgets_count"`
}

// RedashManager struct for RedashPlugin mackerel plugin
type RedashManager struct {
	LastRefreshAt        string       `json:"last_refresh_at"`
	OutdatedQueriesCount uint64       `json:"outdated_queries_count"`
	QueryIds             string       `json:"query_ids"`
	Queues               RedashQueues `json:"queues"`
}

// RedashQueues struct for RedashPlugin mackerel plugin
type RedashQueues struct {
	Queries          RedashQuery `json:"queries"`
	ScheduledQueries RedashQuery `json:"scheduled_queries"`
}

// RedashQuery struct for RedashPlugin mackerel plugin
type RedashQuery struct {
	DataSources string `json:"data_sources"`
	Size        uint64 `json:"size"`
}

// RedashTasks struct for RedashPlugin mackerel plugin
type RedashTasks struct {
	Done       []RedashTask `json:"done"`
	Waiting    []RedashTask `json:"waiting"`
	InProgress []RedashTask `json:"in_progress"`
}

// RedashTask struct for RedashPlugin mackerel plugin
type RedashTask struct {
	Username         string  `json:"username"`
	Retries          uint64  `json:"retries"`
	ScheduledRetries uint64  `json:"scheduled_retries"`
	TaskID           string  `json:"task_id"`
	CreatedAt        float64 `json:"created_at"`
	UpdatedAt        float64 `json:"updated_at"`
	State            string  `json:"state"`
	QueryID          uint64  `json:"query_id"`
	RunTime          float64 `json:"run_time"`
	Error            string  `json:"error"`
	Scheduled        bool    `json:"scheduled"`
	StartedAt        float64 `json:"started_at"`
	DataSourceID     uint64  `json:"data_source_id"`
	QueryHash        string  `json:"query_hash"`
}

// GraphDefinition interface for mackerelplugin
func (p RedashPlugin) GraphDefinition() map[string](mp.Graphs) {
	return map[string](mp.Graphs){
		p.Prefix + ".general": mp.Graphs{
			Label: "re:dash General",
			Unit:  "integer",
			Metrics: [](mp.Metrics){
				mp.Metrics{Name: "dashboards_count", Label: "Dashboards Count", Diff: false, Type: "uint64"},
				mp.Metrics{Name: "queries_count", Label: "Queries Count", Diff: false, Type: "uint64"},
				mp.Metrics{Name: "widgets_count", Label: "Widgets Count", Diff: false, Type: "uint64"},
			},
		},
		p.Prefix + ".query_results": mp.Graphs{
			Label: "re:dash Query Results",
			Unit:  "integer",
			Metrics: [](mp.Metrics){
				mp.Metrics{Name: "query_results_count", Label: "Query Results Count", Diff: false, Type: "uint64"},
				mp.Metrics{Name: "unused_query_results_count", Label: "Unused Query Results Count", Diff: false, Type: "uint64"},
			},
		},
		p.Prefix + ".queues": mp.Graphs{
			Label: "re:dash Queues",
			Unit:  "integer",
			Metrics: [](mp.Metrics){
				mp.Metrics{Name: "queries_size", Label: "Queries", Diff: false, Type: "uint64"},
				mp.Metrics{Name: "scheduled_queries_size", Label: "Scheduled Queries", Diff: false, Type: "uint64"},
			},
		},
		p.Prefix + ".tasks": mp.Graphs{
			Label: "re:dash Tasks",
			Unit:  "integer",
			Metrics: [](mp.Metrics){
				mp.Metrics{Name: "done", Label: "Done", Diff: false, Type: "uint64"},
				mp.Metrics{Name: "in_progress", Label: "In Progress", Diff: false, Type: "uint64"},
				mp.Metrics{Name: "waiting", Label: "Waiting", Diff: false, Type: "uint64"},
			},
		},
	}
}

// FetchMetrics interface for mackerelplugin
func (p RedashPlugin) FetchMetrics() (map[string]interface{}, error) {
	status, err := getStatus(p)
	if err != nil {
		return nil, fmt.Errorf("Faild to fetch re:dash status: %s", err)
	}

	tasks, err := getTasks(p)
	if err != nil {
		return nil, fmt.Errorf("Faild to fetch re:dash tasks: %s", err)
	}

	return map[string]interface{}{
		"dashboards_count":           status.DashboardsCount,
		"queries_count":              status.QueriesCount,
		"query_results_count":        status.QueryResultsCount,
		"unused_query_results_count": status.UnusedQueryResultsCount,
		"widgets_count":              status.WidgetsCount,
		"queries_size":               status.Manager.Queues.Queries.Size,
		"scheduled_queries_size":     status.Manager.Queues.ScheduledQueries.Size,
		"done":        uint64(len(tasks.Done)),
		"in_progress": uint64(len(tasks.InProgress)),
		"waiting":     uint64(len(tasks.Waiting)),
	}, nil
}

func getStatus(p RedashPlugin) (*RedashStatus, error) {
	url := p.URL + "/status.json?api_key=" + p.APIKey

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	timeout := time.Duration(time.Duration(p.Timeout) * time.Second)
	client := &http.Client{
		Timeout: timeout,
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New("HTTP response code is not 200")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var status *RedashStatus
	json.Unmarshal(body, &status)

	return status, nil
}

func getTasks(p RedashPlugin) (*RedashTasks, error) {
	url := p.URL + "/api/admin/queries/tasks?api_key=" + p.APIKey

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	timeout := time.Duration(time.Duration(p.Timeout) * time.Second)
	client := &http.Client{
		Timeout: timeout,
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New("HTTP response code is not 200")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var tasks *RedashTasks
	json.Unmarshal(body, &tasks)

	return tasks, nil
}

func main() {
	optURL := flag.String("url", "http://localhost:5000", "Base URL")
	optAPIKey := flag.String("api-key", "", "API Key")
	optPrefix := flag.String("metric-key-prefix", "redash", "Metric key prefix")
	optTimeout := flag.Uint("timeout", 5, "Timeout")
	optTempfile := flag.String("tempfile", "", "Temp file name")
	flag.Parse()

	if *optAPIKey == "" {
		fmt.Fprintln(os.Stderr, "API Key is required")
		os.Exit(1)
	}

	p := RedashPlugin{
		URL:     *optURL,
		APIKey:  *optAPIKey,
		Prefix:  *optPrefix,
		Timeout: *optTimeout,
	}

	helper := mp.NewMackerelPlugin(p)
	helper.Tempfile = *optTempfile
	if helper.Tempfile == "" {
		helper.Tempfile = fmt.Sprintf("/tmp/mackerel-plugin-%s", *optPrefix)
	}

	helper.Run()
}
