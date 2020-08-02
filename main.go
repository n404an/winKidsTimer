package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

var defaultConfig configFile = configFile{
	Timers: []string{"08:00:00-09:00:00",
		"11:00:00-13:00:00",
		"16:00:00-18:00:00",
		"21:00:00-23:00:00"},
	Msg:            "Я хочу спать!",
	MsgDeny:        "Я не выспался. Посплю ещё",
	WaitingSec:     60,
	WaitingSecDeny: 30,
}

type configFile struct {
	Timers         []string
	Msg            string
	MsgDeny        string
	WaitingSec     int
	WaitingSecDeny int
}

type timers map[time.Time]time.Time

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	dir := ".winKidTimer"
	fileName := "config.yml"
	path2dir := filepath.Join(homeDir, dir)
	path2file := filepath.Join(homeDir, dir, fileName)

	if f, err := os.Stat(path2file); os.IsNotExist(err) {
		fmt.Println("нету", err)
		createConfigFile(path2dir, path2file)
	} else {
		fmt.Println("есть", f, f.Mode())
	}
	file := &configFile{}

	if f, err := ioutil.ReadFile(path2file); err != nil {
		fmt.Println("Не смог", err)
	} else {
		if err := yaml.Unmarshal(f, file); err != nil {
			fmt.Println("Не смог", err)
		}
	}
	fmt.Println("asd", file)
	timers := parseTimers(file)
	file.Msg = strings.ReplaceAll(file.Msg, " ", "_")
	file.MsgDeny = strings.ReplaceAll(file.MsgDeny, " ", "_")
	shutDown(file, timers)
}

func shutDown(f *configFile, t timers) {
	now := time.Now()
	var canPlay bool
	for i, v := range t {
		fmt.Println("timers: ", i, ": ", v, now)
		if i.Before(now) && v.After(now) {
			canPlay = true
			break
		}
	}
	args := new(strings.Builder)
	args.WriteString("shutdown -s -t ")
	args.WriteString(strconv.Itoa(f.WaitingSecDeny))
	args.WriteString(" -c \"")
	fmt.Println(now)
	if !canPlay {
		nextShutdown := getDuration(canPlay, now, t)
		fmt.Println(nextShutdown)
		args.WriteString(f.MsgDeny)
		args.WriteString("_")
		args.WriteString(strconv.Itoa(nextShutdown / 60))
		args.WriteString("_мин.\"")

		if err := exec.Command("cmd", "/C", args.String()).Run(); err != nil {
			fmt.Println("Failed to initiate shutdown:", err)
		}
	} else {
		args.WriteString(f.Msg)
		args.WriteString("\"")
		fmt.Println(args.String())
		nextShutdown := getDuration(canPlay, now, t)
		fmt.Println(nextShutdown)
		time.Sleep(time.Duration(nextShutdown) * time.Second)
		if err := exec.Command("cmd", "/C", args.String()).Run(); err != nil {
			fmt.Println("Failed to initiate shutdown:", err)
		}
	}

}
func getDuration(canPlay bool, now time.Time, t timers) int {

	next := int(now.Unix())

	sortTimers := make([]int, 0, len(t))

	if canPlay {
		for _, i := range t {
			sortTimers = append(sortTimers, int(i.Unix()))
		}
	} else {
		for i := range t {
			sortTimers = append(sortTimers, int(i.Unix()))
		}
	}

	fmt.Println(sortTimers, "now: ", now.Unix())
	sort.Ints(sortTimers)
	fmt.Println(sortTimers, "now: ", now.Unix())

	for _, t := range sortTimers {
		if next < t {
			fmt.Println("min -t: ", time.Unix(int64(next), 0), time.Unix(int64(t), 0), t-next, time.Duration(t-next)*time.Second)
			next = t - next
			break
		}
	}
	return next
}
func parseTimers(c *configFile) timers {
	t := make(timers, len(c.Timers))
	now := time.Now()
	for _, v := range c.Timers {
		v = strings.ReplaceAll(v, " ", "")
		fromTo := strings.Split(v, "-")

		from := parseTime(fromTo[0], now)
		to := parseTime(fromTo[1], now)
		t[from] = to

		from = parseTime(fromTo[0], now.Add(24*time.Hour))
		to = parseTime(fromTo[1], now.Add(24*time.Hour))
		t[from] = to
	}

	return t
}

func parseTime(s string, now time.Time) time.Time {
	a := strings.Split(s, ":")
	hh, err := strconv.ParseInt(a[0], 10, 64)
	if err != nil {
		fmt.Println("parse hh: ", err)
	}
	mm, err := strconv.ParseInt(a[1], 10, 64)
	if err != nil {
		fmt.Println("parse mm: ", err)
	}
	ss, err := strconv.ParseInt(a[2], 10, 64)
	if err != nil {
		fmt.Println("parse ss: ", err)
	}

	return time.Date(now.Year(), now.Month(), now.Day(), int(hh), int(mm), int(ss), 0, now.Location())
}

func createConfigFile(path2dir, path2file string) {

	if d, err := os.Stat(path2dir); os.IsNotExist(err) {
		fmt.Println("нету папки", err)
		if err = os.Mkdir(path2dir, 0777); err != nil {
			fmt.Println("не смог создать папку ", err)
			return
		} else {
			fmt.Println("создал папку", err)
		}
	} else {
		fmt.Println("есть папка", d, d.IsDir(), d.Mode())
	}
	if y, err := yaml.Marshal(defaultConfig); err != nil {
		fmt.Println("Не смог", err)
	} else {
		fmt.Println(y)
		if file, err := os.Create(path2file); err != nil {
			fmt.Println("Не смог", err)
		} else {
			if _, err := file.Write(y); err != nil {
				fmt.Println("Не смог", err)

			}
			file.Close()
		}
	}
}
