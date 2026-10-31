package main

import (
	"testing"
	"time"
)

func TestNewServerTimeouts(t *testing.T) {
	server := newServer()
	for _, tt := range []struct {
		name    string
		timeout time.Duration
	}{
		{"ReadHeaderTimeout", server.ReadHeaderTimeout},
		{"ReadTimeout", server.ReadTimeout},
		{"WriteTimeout", server.WriteTimeout},
		{"IdleTimeout", server.IdleTimeout},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if tt.timeout <= 0 {
				t.Errorf("%s = %v, want a positive timeout", tt.name, tt.timeout)
			}
		})
	}
}
