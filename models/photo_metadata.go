package models

type TimeInfo struct {
	Timestamp string `json:"timestamp"`
	Formatted string `json:"formatted"`
}

type PhotoMetadata struct {
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	CreationTime   TimeInfo  `json:"creationTime"`
	PhotoTakenTime *TimeInfo `json:"photoTakenTime,omitempty"`
}
