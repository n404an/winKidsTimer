package main

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"log"
	"os"
	"time"
)

var tsLogs *logs

type logs struct{ Ts []string }

func getLogs() (lastTs time.Time, err error) {
	tsLogs = &logs{}
	lastTs = now
	data, err := os.ReadFile(logsPath)
	if err != nil {
		return lastTs, fmt.Errorf("Логи не обнаружены. %v\n", err)
	}

	if err = jsoniter.Unmarshal(data, tsLogs); err != nil {
		return lastTs, fmt.Errorf("Битые логи. %v\n", err)
	}
	if len(tsLogs.Ts) == 0 {
		return lastTs, fmt.Errorf("Пустые логи\n")
	}
	return time.Parse(time.DateTime, tsLogs.Ts[len(tsLogs.Ts)-1])
}

func (l *logs) add() {
	l.Ts = append(l.Ts, now.Format(time.DateTime))
	data, err := jsoniter.Marshal(l)
	if err != nil {
		log.Println("(l *logs) add()", err)
		return
	}
	fmt.Println("(l *logs) add()", l, string(data))
	if err := os.WriteFile(logsPath, data, 0644); err != nil {
		log.Println("os.WriteFile", err)
	}
}
