package cyber

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
