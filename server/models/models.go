package models

import (
	"log"
)

type Config struct {
	GeoSystemID    string
	CalorificValue float64
}

type Env struct {
	Config         Config
	Logger         *log.Logger
	GeoUser        string
	GeoPass        string
	EnableAPI      bool
	EnableInfluxDB bool
	DebugMode      bool
	APIKey         string
}

type LiveUsageData struct {
	Watts       float64 `json:"watts"`
	LastUpdated int64   `json:"lastUpdated"`
}
type LiveUsage struct {
	Electricity LiveUsageData `json:"electricity"`
	Gas         LiveUsageData `json:"gas"`
}

type PeriodicUsageData struct {
	ReadingTime      int64   `json:"readingTime"`
	TotalConsumption float64 `json:"totalConsumption"`
	Unit             string  `json:"unit"`
}
type PeriodicUsage struct {
	Electricity PeriodicUsageData `json:"electricity"`
	Gas         PeriodicUsageData `json:"gas"`
}
