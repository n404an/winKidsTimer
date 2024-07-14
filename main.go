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
		log.Println("timeToShutdown>0 осталось", timeToShutdown.String())
		shutDown(int(timeToShutdown.Seconds()), cfg.Msg)
	case nextPlayTime < 0:
		tsLogs.add()
		log.Println("nextPlayTime < 0 осталось", timeToShutdown.String())
		shutDown(int(timeToShutdown.Seconds()), cfg.MsgDeny)
	case timeToShutdown < 0:
		log.Println("timeToShutdown<0")
		shutDown(cfg.WaitingSec, cfg.Msg)
	}
}

func shutDown(wait int, msg string) {
	if devMode {
		return
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

func setup() {
	/*	mode := 0
		dynMode:
			fmt.Println("Какой режим работы?")
			fmt.Println("1. Статический(фиксированный график работы ПК)  (НЕ работает)")
			fmt.Println("2. Динамический(например, с 7ч до 21ч, не дольше часа, с перерывами минимум 2 часа)")
			fmt.Println("3. Выходим.")
			fmt.Fscan(os.Stdin, &mode)

			switch mode {
			case 1, 2:
				cfg.DynamicMode = 2 // TODO mode 1
			case 3:
				fmt.Println("\nПока.")
				return
			default:
				fmt.Println("\nПопробуем ещё раз...")
				goto dynMode
			}
			fmt.Println(" >>>>> Выбран режим", mode)
			fmt.Println("\nТекущие настройки:")

			cfg.Print()

		saveLabel:
			fmt.Println("\nПрименить настройки?")
			fmt.Println("1. Да.")
			fmt.Println("2. Редактировать текущие.")
			fmt.Println("3. Нет и начать заново.")
			fmt.Println("4. Нет и выйти.")
			fmt.Fscan(os.Stdin, &mode)

			switch mode {
			case 1:
				// копировать exe в автозагрузку
				// %AppData%\Microsoft\Windows\Start Menu\Programs\Startup
				//C:\Users\omfgw\AppData\Roaming\Microsoft\Windows\Start Menu\Programs\Startup
				pCfg.Set("cfg", cfg)
				outputFile := filepath.Join(usrStartup, processName)

				if _, err := os.Stat(filepath.Join(usrStartup, processName)); err != nil {
					if err := copyFile(progName, outputFile); err != nil {
						log.Println(err)
					} else {
						fmt.Println("Файл скопирован в автозагрузку.")
					}
				}
				fmt.Println("\nГотово. При следующем запуске системы программа начнёт работать.")
			case 2:
				modeEdit := 0
			editLabel:
				fmt.Println("\nЧто будем редактировать?")
				fmt.Println("1. Интервал")
				fmt.Println("2. Время работы")
				fmt.Println("3. Время между включениями")
				fmt.Println("4. Вернуться назад")

				fmt.Fscan(os.Stdin, &modeEdit)
				switch modeEdit {
				case 1:
					fmt.Print("\nТекущий интервал: ")
					v := cfg.Timers[0]
					fmt.Printf("%s:%s-%s:%s\n", getTime(v[0]), getTime(v[1]), getTime(v[2]), getTime(v[3]))
					fmt.Println("В таком же формате(ЧЧ:ММ-ЧЧ:ММ) опишите нужный интервал.")
					var hh, mm, hh2, mm2 int
					_, err := fmt.Scanf("\n%d:%d-%d:%d", &hh, &mm, &hh2, &mm2)
					if err != nil {
						log.Println(err)
					}
					cfg.Timers = [][4]int{[4]int{hh, mm, hh2, mm2}}
					fmt.Println(hh, mm, hh2, mm2)
					fmt.Printf("Готово. Установлен интервал %d:%d-%d:%d\n", hh, mm, hh2, mm2)
					goto editLabel
				case 2:
					fmt.Print("\nТекущее время работы: ")
					v := cfg.MaxPlay
					if v >= 60 {
						v /= 60
						fmt.Println(v, "мин.")
					} else {
						fmt.Println(v, "сек.")
					}
					n := 0
					fmt.Println("Напишите нужное время в минутах.")
					_, err := fmt.Scanf("\n%d", &n)
					if err != nil {
						log.Println(err)
					}
					cfg.MaxPlay = n * 60
					fmt.Printf("Готово. Установлено время работы: %d мин.\n", n)
					goto editLabel
				case 3:
					fmt.Print("\nТекущее время между включениями: ")
					v := cfg.PlayInterval
					if v >= 60 {
						v /= 60
						fmt.Println(v, "мин.")
					} else {
						fmt.Println(v, "сек.")
					}
					n := 0
					fmt.Println("Напишите нужное время в минутах.")
					_, err := fmt.Scanf("\n%d", &n)
					if err != nil {
						log.Println(err)
					}
					cfg.PlayInterval = n * 60
					fmt.Printf("Готово. Установлено время между включениями: %d мин.\n", n)
					goto editLabel
				case 4:
					goto saveLabel
				default:
					fmt.Println("\nПопробуем ещё раз...")
					goto editLabel
				}
			case 3:
				mode = 0
				goto dynMode
			case 4:
				return
			default:
				fmt.Println("\nПопробуем ещё раз...")
				goto saveLabel
			}*/
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
