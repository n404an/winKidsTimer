package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	progName                  = "winKidsTimer.exe"
	processName               = "winKidsTimerProc.exe"
	usrStartup                string
	processPath               string
	exPath, cfgPath, logsPath string
	cfg                       *config

	now time.Time

	devMode = false
)

func init() {
	homeDir, _ := os.UserHomeDir()
	usrStartup = filepath.Join(homeDir, "AppData", "Roaming",
		"Microsoft", "Windows", "Start Menu", "Programs", "Startup")

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath = filepath.Dir(ex)
	dir := ".winKidsTimer"

	cfgPath = filepath.Join(homeDir, dir, "cfg.json")
	logsPath = filepath.Join(homeDir, dir, "logs.json")
	processPath = filepath.Join(usrStartup, processName)

	if _, err := os.Stat(filepath.Join(homeDir, dir)); err != nil {
		log.Println(err)
		if err := os.Mkdir(filepath.Join(homeDir, dir), os.ModeDir); err != nil {
			log.Println(err)
		}
	}

	now = time.Now()
	_, zone := now.Zone()
	now = now.UTC().Add(time.Duration(int64(zone)) * time.Second).Truncate(time.Second)
	cfg = &config{}
	cfg.init()
}

func main() {
	_, err := os.Stat(processPath)
	switch {
	case err != nil:
		fmt.Println("отсутствует в автозагрузке")

		if err := copyFile(progName, processPath); err != nil {
			log.Println(err)
		} else {
			fmt.Println("Файл скопирован в автозагрузку.")
		}
		// setup()
		run()
	case usrStartup == exPath:
		fmt.Println("запуск из автозагрузки")
		run()
	default:
		fmt.Println("редактирование")
		// edit()
		if err := copyFile(progName, processPath); err != nil {
			log.Println(err)
		} else {
			fmt.Println("Файл скопирован в автозагрузку.")
		}
	}
}

func run() {
	maxPlay, err := time.ParseDuration(cfg.MaxPlay)
	if err != nil {
		log.Println(err)
		return
	}
	chill, err := time.ParseDuration(cfg.PlayInterval)
	if err != nil {
		log.Println(err)
		return
	}
	durFrom, err := time.ParseDuration(cfg.From)
	if err != nil {
		log.Println(err)
		return
	}
	durTo, err := time.ParseDuration(cfg.To)
	if err != nil {
		log.Println(err)
		return
	}

	start := now.Truncate(time.Hour * 24).Add(durFrom)
	end := now.Truncate(time.Hour * 24).Add(durTo)

	lastTs, err := getLogs()
	if err != nil {
		log.Println(err)
		tsLogs.add()
	}

	log.Println("run now  ", now)
	log.Println("run start", start)
	log.Println("run end  ", end)
	log.Println("run play ", maxPlay)
	log.Println("run last ", lastTs)

	timeToShutdown := lastTs.Add(maxPlay).Sub(now) // now.Sub(lastTs.Add(maxPlay))
	if maxTimeToPlay := end.Sub(now); timeToShutdown > maxTimeToPlay {
		timeToShutdown = maxTimeToPlay
	}

	nextPlayTime := timeToShutdown + chill
	log.Println("run timeToShutdown", timeToShutdown)
	log.Println("run nextPlayTime  ", nextPlayTime)

	switch {
	case now.After(end) || now.Before(start):
		log.Println("now.After(end) && now.Before(start)")
		shutDown(cfg.WaitingSec, cfg.Msg)
	case timeToShutdown > 0:
		log.Println("timeToShutdown>0 осталось", timeToShutdown.String(), timeToShutdown.Seconds())
		shutDown(int(timeToShutdown.Seconds()), cfg.Msg)
	case nextPlayTime < 0:
		tsLogs.add()
		lastTs, _ = getLogs()
		timeToShutdown = lastTs.Add(maxPlay).Sub(now)
		if maxTimeToPlay := end.Sub(now); timeToShutdown > maxTimeToPlay {
			timeToShutdown = maxTimeToPlay
		}
		nextPlayTime = timeToShutdown + chill
		log.Println("run last ", lastTs)
		log.Println("run timeToShutdown", timeToShutdown)
		log.Println("run nextPlayTime  ", nextPlayTime)
		log.Println("nextPlayTime < 0 осталось", timeToShutdown.String())
		shutDown(int(timeToShutdown.Seconds()), cfg.Msg)
	case timeToShutdown < 0:
		log.Println("timeToShutdown<0")
		shutDown(cfg.WaitingSec, cfg.Msg)
	}
}

func shutDown(wait int, msg string) {
	if devMode {
		return
	}
	if wait <= 0 {
		wait = cfg.WaitingSec
	}
	msg = strings.ReplaceAll(msg, " ", "_")
	m := fmt.Sprintf("shutdown -s -t %d -c \"%v\"", wait, msg)

	if err := exec.Command("cmd", "/C", m).Run(); err != nil {
		fmt.Println("Failed to initiate shutdown:", err)
	}
}

func shutDownAbort() {
	c := `shutdown -a`
	if err := exec.Command("cmd", "/C", c).Run(); err != nil {
		if msg, ok := err.(*exec.ExitError); ok {
			if msg.ExitCode() == 1116 {
				fmt.Println("\nПрограмма не запущена")
				return
			}
		}
		fmt.Println("Failed to initiate shutdown:", err)
		return
	}
	fmt.Println("Отключение отменено")
}
func taskkill() {
	c := `taskkill /f /t /im ` + processName
	if err := exec.Command("cmd", "/C", c).Run(); err != nil {
		fmt.Println("Failed to initiate taskkill:", err)
	}
}
func delFromStartup(usrStartup string) {
	if err := os.Remove(filepath.Join(usrStartup, processName)); err != nil {
		log.Println(err)
		return
	}
	fmt.Println("Готово.")
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return nil
}
