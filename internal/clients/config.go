package clients

const RetryPolicy = `{
		"methodConfig": [{
			"name": [
				{"service": "report.Report"},
				{"service": "schedule.Schedule"}
			],
			"retryPolicy": {
				"MaxAttempts": 10,
				"InitialBackoff": "1s",
				"MaxBackoff": "60s",
				"BackoffMultiplier": 2,
				"RetryableStatusCodes": [ "UNAVAILABLE" ]
		}
	}]}`
