// Package helpers provides helper functions
package helpers

import "fmt"

func GenerateExampleConfig() {
	newConfig := `
[
	{
		"name": "Workflow 1",
		"enabled": true,
		"trigger": { "every": "1d", "beginAt": "10:00" },
		"onFailure": "continue",
		"steps": [
			{
				"name": "step 1",
				"program": "/usr/bin/some_program",
				"args": [
					"--verbose",
					"--file",
					"some_file_name"
				],
				"timeout": "11s",
				"pause": "3s"
			},
			{
				"name": "step 2",
				"program": "/usr/bin/some_program",
				"args": [],
				"pause": "0s"
			}
		]
	},
	{
		"name": "Workflow 2",
		"enabled": true,
		"trigger": { "every": "1d", "beginAt": "10:35" },
		"onFailure": "abort",
		"steps": [
			{
				"name": "daily",
				"program": "/opt/sbin/some_script",
				"args": [
					"--sleep",
					"5"
				]
			}
		]
	},
	{
		"name": "Workflow 3",
		"enabled": false,
		"disabledReason": "demonstrating enabled features",
		"trigger": { "every": "1d", "beginAt": "10:35" },
		"onFailure": "retry",
		"retry": {
			"numberRetries": 3,
			"pauseSeconds": 1
		},
		"steps": [
			{
				"name": "daily",
				"program": "/opt/sbin/some_script",
				"args": [
					"--sleep",
					"5"
				]
			}
		]
	}
]
`

	fmt.Println(newConfig)
}
