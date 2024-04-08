package main

import (
	"time"
)

type AppConfig struct {
	ServerPort   string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

const (
	defaultServerPort   = "8080"
	defaultReadTimeout  = 1 * time.Minute
	defaultWriteTimeout = 1 * time.Minute
)

func getValueWithDefault(value string, defaultVal string) string {
	if value != "" {
		return value
	}
	return defaultVal
}
func getDurationWithDefault(value time.Duration, defaultVal time.Duration) time.Duration {
	if value == 0 {
		return defaultVal
	}
	return value
}
