package main

import (
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

const APIKey = "test"

func TestGetStatus(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// https://github.com/getredash/redash/blob/master/docs/usage/maintenance.rst#monitoring
	jsonStr := `{
	  "dashboards_count": 30,
	  "manager": {
	    "last_refresh_at": "1465392784.433638",
	    "outdated_queries_count": 1,
	    "query_ids": "[34]",
	    "queues": {
	      "queries": {
	        "data_sources": "Redshift data, re:dash metadata, MySQL data, MySQL read-only, Redshift read-only",
	        "size": 1
	      },
	      "scheduled_queries": {
	        "data_sources": "Redshift data, re:dash metadata, MySQL data, MySQL read-only, Redshift read-only",
	        "size": 0
	      }
	    }
	  },
	  "queries_count": 204,
	  "query_results_count": 11161,
	  "redis_used_memory": "6.09M",
	  "unused_query_results_count": 32,
	  "version": "0.10.0+b1774",
	  "widgets_count": 176,
	  "workers": []
	}`

	url := "http://httpmock/status.json?api_key=" + APIKey

	httpmock.RegisterResponder("GET", url,
		httpmock.NewStringResponder(200, jsonStr))

	p := RedashPlugin{
		URL:     "http://httpmock",
		APIKey:  APIKey,
		Prefix:  "redash",
		Timeout: 5,
	}
	status, _ := getStatus(p)

	assert.EqualValues(t, 30, status.DashboardsCount)
	assert.EqualValues(t, 204, status.QueriesCount)
	assert.EqualValues(t, 11161, status.QueryResultsCount)
	assert.EqualValues(t, "6.09M", status.RedisUsedMemory)
	assert.EqualValues(t, 32, status.UnusedQueryResultsCount)
	assert.EqualValues(t, "0.10.0+b1774", status.Version)
	assert.EqualValues(t, 176, status.WidgetsCount)

	assert.EqualValues(t, "1465392784.433638", status.Manager.LastRefreshAt)
	assert.EqualValues(t, 1, status.Manager.OutdatedQueriesCount)
	assert.EqualValues(t, "[34]", status.Manager.QueryIds)
	assert.EqualValues(t, "Redshift data, re:dash metadata, MySQL data, MySQL read-only, Redshift read-only", status.Manager.Queues.Queries.DataSources)
	assert.EqualValues(t, 1, status.Manager.Queues.Queries.Size)
	assert.EqualValues(t, "Redshift data, re:dash metadata, MySQL data, MySQL read-only, Redshift read-only", status.Manager.Queues.ScheduledQueries.DataSources)
	assert.EqualValues(t, 0, status.Manager.Queues.ScheduledQueries.Size)

	httpmock.Reset()
	httpmock.RegisterResponder("GET", url,
		httpmock.NewStringResponder(500, ""))

	_, err := getStatus(p)
	assert.Equal(t, "HTTP response code is not 200", err.Error())
}

func TestGetTasks(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	jsonStr := `{
	    "waiting": [],
	    "done": [{
	        "username": "Scheduled",
	        "retries": 0,
	        "scheduled_retries": 0,
	        "task_id": "8d478084-efeb-4919-9e35-c9888b92b181",
	        "created_at": 1467452246.586044,
	        "updated_at": 1467452247.775033,
	        "state": "finished",
	        "query_id": 16,
	        "run_time": 1.1670680046081543,
	        "error": null,
	        "scheduled": true,
	        "started_at": 1467452246.593749,
	        "data_source_id": 1,
	        "query_hash": "97b9b061952cc1af31040ce7a04bedd1"
	    }],
	    "in_progress": [{
	        "username": "admin",
	        "retries": 0,
	        "started_at": null,
	        "task_id": "c432d243-ca17-4b08-b5c3-752e666722ee",
	        "created_at": 1467459544.887608,
	        "updated_at": 1467459544.887616,
	        "state": "created",
	        "query_id": 32,
	        "run_time": null,
	        "scheduled": false,
	        "scheduled_retries": 0,
	        "data_source_id": 1,
	        "query_hash": "3bbdb1e5ec27585b3ef5e2601eace77b"
	    }]
	}`

	url := "http://httpmock/api/admin/queries/tasks?api_key=" + APIKey

	httpmock.RegisterResponder("GET", url,
		httpmock.NewStringResponder(200, jsonStr))

	p := RedashPlugin{
		URL:     "http://httpmock",
		APIKey:  APIKey,
		Prefix:  "redash",
		Timeout: 5,
	}
	tasks, _ := getTasks(p)

	assert.EqualValues(t, 1, len(tasks.Done))
	assert.EqualValues(t, 0, len(tasks.Waiting))
	assert.EqualValues(t, 1, len(tasks.InProgress))

	assert.EqualValues(t, "admin", tasks.InProgress[0].Username)
	assert.EqualValues(t, 0, tasks.InProgress[0].Retries)
	assert.EqualValues(t, 0, tasks.InProgress[0].StartedAt)
	assert.EqualValues(t, "c432d243-ca17-4b08-b5c3-752e666722ee", tasks.InProgress[0].TaskID)
	assert.EqualValues(t, 1467459544.887608, tasks.InProgress[0].CreatedAt)
	assert.EqualValues(t, 1467459544.887616, tasks.InProgress[0].UpdatedAt)
	assert.EqualValues(t, "created", tasks.InProgress[0].State)
	assert.EqualValues(t, 32, tasks.InProgress[0].QueryID)
	assert.EqualValues(t, 0, tasks.InProgress[0].RunTime)
	assert.EqualValues(t, false, tasks.InProgress[0].Scheduled)
	assert.EqualValues(t, 0, tasks.InProgress[0].ScheduledRetries)
	assert.EqualValues(t, 1, tasks.InProgress[0].DataSourceID)
	assert.EqualValues(t, "3bbdb1e5ec27585b3ef5e2601eace77b", tasks.InProgress[0].QueryHash)

	httpmock.Reset()
	httpmock.RegisterResponder("GET", url,
		httpmock.NewStringResponder(500, ""))

	_, err := getTasks(p)
	assert.Equal(t, "HTTP response code is not 200", err.Error())
}
