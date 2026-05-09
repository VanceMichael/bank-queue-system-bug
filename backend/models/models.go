package models

import (
	"time"
)

type BusinessType string

const (
	BusinessTypePersonal BusinessType = "personal"
	BusinessTypeCorporate BusinessType = "corporate"
	BusinessTypeVIP BusinessType = "vip"
)

type QueueStatus string

const (
	QueueStatusWaiting QueueStatus = "waiting"
	QueueStatusCalling QueueStatus = "calling"
	QueueStatusProcessing QueueStatus = "processing"
	QueueStatusCompleted QueueStatus = "completed"
	QueueStatusMissed QueueStatus = "missed"
)

type WindowStatus string

const (
	WindowStatusOpen WindowStatus = "open"
	WindowStatusClosed WindowStatus = "closed"
	WindowStatusPaused WindowStatus = "paused"
)

type QueueNumber struct {
	ID           string       `json:"id"`
	Number       string       `json:"number"`
	BusinessType BusinessType `json:"businessType"`
	Status       QueueStatus  `json:"status"`
	WindowID     string       `json:"windowId,omitempty"`
	CreatedAt    time.Time    `json:"createdAt"`
	CalledAt     *time.Time   `json:"calledAt,omitempty"`
	CompletedAt  *time.Time   `json:"completedAt,omitempty"`
	IsVIP        bool         `json:"isVip"`
	Priority     int          `json:"priority"`
}

type Window struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Status         WindowStatus   `json:"status"`
	BusinessTypes  []BusinessType `json:"businessTypes"`
	CurrentQueue   *QueueNumber   `json:"currentQueue,omitempty"`
	LastActiveAt   *time.Time     `json:"lastActiveAt,omitempty"`
}

type BusinessTypeConfig struct {
	Type        BusinessType `json:"type"`
	Name        string       `json:"name"`
	Prefix      string       `json:"prefix"`
	AverageTime int          `json:"averageTime"`
	Description string       `json:"description"`
	IsActive    bool         `json:"isActive"`
}

type Statistics struct {
	TotalCustomers      int64                              `json:"totalCustomers"`
	AverageWaitTime     float64                            `json:"averageWaitTime"`
	AverageProcessTime  float64                            `json:"averageProcessTime"`
	WindowEfficiency    map[string]float64                 `json:"windowEfficiency"`
	PeakHours           map[int]int64                      `json:"peakHours"`
	BusinessTypeStats   map[BusinessType]BusinessTypeStat  `json:"businessTypeStats"`
}

type BusinessTypeStat struct {
	Count          int64   `json:"count"`
	AverageWait    float64 `json:"averageWait"`
	AverageProcess float64 `json:"averageProcess"`
}
