package main

import "time"

type Response struct {
	StatusCode int         `json:"statusCode"`
	Body       interface{} `json:"body"`
}

type EventMetadata struct {
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
	CreatedAt time.Time `json:"created_at"`
	CloudID   string    `json:"cloud_id"`
	FolderID  string    `json:"folder_id"`
}

type Details struct {
	TriggerID string `json:"trigger_id"`
}

type TimerTriggerMessage struct {
	EventMetadata EventMetadata `json:"event_metadata"`
	Details       Details       `json:"details"`
}

type TimerTriggerEvent struct {
	Messages []TimerTriggerMessage `json:"messages"`
}
