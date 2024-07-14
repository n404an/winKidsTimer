package main

import (
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	devMode = true
	now = now.Add(5 * time.Hour)
	run()
}
