package main

import "testing"

func TestCalculateMetricsCase1(t *testing.T) {
	historyData := `
	{
		"history": [
			{
				"before": {
					"status": "in definition pm"
				},
				"after": {
					"status": "in definition dev"
				},
				"date": "1686578015783"
			},
			{
				"before": {
					"status": "in definition dev"
				},
				"after": {
					"status": "in development"
				},
				"date": "1687785838064"
			},
			{
				"before": {
					"status": "in development"
				},
				"after": {
					"status": "blocked"
				},
				"date": "1687995461000"
			},
			{
				"before": {
					"status": "blocked"
				},
				"after": {
					"status": "in development"
				},
				"date": "1688168261000"
			},
			{
				"before": {
					"status": "in development"
				},
				"after": {
					"status": "completado"
				},
				"date": "1688649601245"
			}
		]
	}`

	history := parserJSON([]byte(historyData))
	taskInfo := TaskInfo{
		TaskHeaderData: TaskHeaderData{
			Id:        "85zt8cyjd",
			StartDate: "1685749061000",
		},
		History: history,
	}

	metricsPerState := calculateTimePerState(&taskInfo)

	metrics := calculateMetrics(metricsPerState)

	expectedLeadTime := 34
	expectedCycleTime := 10
	expectedBlockedTime := 2
	expectedFlowEfficiency := 80.00

	if metrics.LeadTime != expectedLeadTime {
		t.Errorf("Lead Time incorrecto, se esperaba %d pero se obtuvo %d", expectedLeadTime, metrics.LeadTime)
	}

	if metrics.CycleTime != expectedCycleTime {
		t.Errorf("Cycle Time incorrecto, se esperaba %d pero se obtuvo %d", expectedCycleTime, metrics.CycleTime)
	}

	if metrics.BlockedTime != expectedBlockedTime {
		t.Errorf("Blocked Time incorrecto, se esperaba %d pero se obtuvo %d", expectedBlockedTime, metrics.BlockedTime)
	}

	if metrics.FlowEfficiency != expectedFlowEfficiency {
		t.Errorf("Flow Efficiency incorrecto, se esperaba %.2f pero se obtuvo %.2f", expectedFlowEfficiency, metrics.FlowEfficiency)
	}
}

func TestCalculateMetricsCase2(t *testing.T) {
	historyData := `
	{
		"history":[
		   {
			  "date":"1683904734343",
			  "before":{
				 "status":"in definition pm",
				 "color":"#EC7010",
				 "orderindex":1,
				 "type":"custom"
			  },
			  "after":{
				 "status":"in design",
				 "color":"#EA80FC",
				 "orderindex":2,
				 "type":"custom"
			  }
		   },
		   {
			  "date":"1684260433287",
			  "before":{
				 "status":"in design",
				 "color":"#EA80FC",
				 "orderindex":2,
				 "type":"custom"
			  },
			  "after":{
				 "status":"in definition dev",
				 "color":"#3397dd",
				 "orderindex":3,
				 "type":"custom"
			  }
		   },
		   {
			  "date":"1685716106015",
			  "before":{
				 "status":"in definition dev",
				 "color":"#3397dd",
				 "orderindex":3,
				 "type":"custom"
			  },
			  "after":{
				 "status":"to develop",
				 "color":"#b5bcc2",
				 "orderindex":4,
				 "type":"custom"
			  }
		   },
		   {
			  "date":"1685969456276",
			  "before":{
				 "status":"to develop",
				 "color":"#b5bcc2",
				 "orderindex":4,
				 "type":"custom"
			  },
			  "after":{
				 "status":"in development",
				 "color":"#81B1FF",
				 "orderindex":5,
				 "type":"custom"
			  }
		   },
		   {
			  "date":"1686263668162",
			  "before":{
				 "status":"in development",
				 "color":"#81B1FF",
				 "orderindex":5,
				 "type":"custom"
			  },
			  "after":{
				 "status":"ready to deploy",
				 "color":"#1bbc9c",
				 "orderindex":8,
				 "type":"custom"
			  }
		   },
		   {
			  "date":"1686310887653",
			  "before":{
				 "status":"ready to deploy",
				 "color":"#1bbc9c",
				 "orderindex":8,
				 "type":"custom"
			  },
			  "after":{
				 "status":"in development",
				 "color":"#81B1FF",
				 "orderindex":5,
				 "type":"custom"
			  }
		   },
		   {
			  "date":"1686599456973",
			  "before":{
				 "status":"in development",
				 "color":"#81B1FF",
				 "orderindex":5,
				 "type":"custom"
			  },
			  "after":{
				 "status":"ready to deploy",
				 "color":"#1bbc9c",
				 "orderindex":8,
				 "type":"custom"
			  }
		   },
		   {
			  "date":"1686660517165",
			  "before":{
				 "status":"ready to deploy",
				 "color":"#1bbc9c",
				 "orderindex":8,
				 "type":"custom"
			  },
			  "after":{
				 "status":"completado",
				 "color":"#6bc950",
				 "orderindex":10,
				 "type":"closed"
			  }
		   }
		]
	 }`

	history := parserJSON([]byte(historyData))
	taskInfo := TaskInfo{
		TaskHeaderData: TaskHeaderData{
			Id:        "85zrzu15w",
			StartDate: "1683874800000",
		},
		History: history,
	}

	metricsPerState := calculateTimePerState(&taskInfo)

	metrics := calculateMetrics(metricsPerState)

	expectedLeadTime := 32
	expectedCycleTime := 8
	expectedBlockedTime := 3
	expectedFlowEfficiency := 62.50

	if metrics.LeadTime != expectedLeadTime {
		t.Errorf("Lead Time incorrecto, se esperaba %d pero se obtuvo %d", expectedLeadTime, metrics.LeadTime)
	}

	if metrics.CycleTime != expectedCycleTime {
		t.Errorf("Cycle Time incorrecto, se esperaba %d pero se obtuvo %d", expectedCycleTime, metrics.CycleTime)
	}

	if metrics.BlockedTime != expectedBlockedTime {
		t.Errorf("Blocked Time incorrecto, se esperaba %d pero se obtuvo %d", expectedBlockedTime, metrics.BlockedTime)
	}

	if metrics.FlowEfficiency != expectedFlowEfficiency {
		t.Errorf("Flow Efficiency incorrecto, se esperaba %.2f pero se obtuvo %.2f", expectedFlowEfficiency, metrics.FlowEfficiency)
	}
}
