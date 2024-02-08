package model

import "time"

type MetricItem struct {
	Namespace  string
	Deployment string
	Kind       string
	Time       time.Time
	Value      float64
}

type MetricFiels struct {
	Request float64 `json:"request"`
	Limit   float64 `json:"limit"`
	Usage   float64 `json:"usage"`
}

type Resources struct {
	Replicas float64     `json:"replicas"`
	Memory   MetricFiels `json:"memory"`
	CPU      MetricFiels `json:"cpu"`
}

type Metrics struct {
	Namespace  string `json:"namespace"`
	Deployment string `json:"deployment"`
	TimeSeries map[time.Time]Resources
}
