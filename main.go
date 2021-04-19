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

	"github.com/recoilme/pudge"
)

var defaultConfig config = config{
	Timers: [][4]int{
		{8, 0, 9, 0},
		{11, 0, 13, 0},
		{16, 0, 18, 0},
		{21, 0, 23, 0},
	},
	Msg:            "Я хочу спать!",
	MsgDeny:        "Я не выспался. Посплю ещё",
	WaitingSec:     60,
	WaitingSecDeny: 30,
	DynamicMode:    0,
	MaxPlay:        3600,
	PlayInterval:   3600 * 2,
}

type config struct {
	Timers         [][4]int
	Msg            string
	MsgDeny        string
	WaitingSec     int
	WaitingSecDeny int
	DynamicMode    int
	MaxPlay        int
	PlayInterval   int
}

type timers map[time.Time]time.Time

var progName = "winKidsTimer.exe"
var processName = "winKidsTimerProc.exe"

func main() {
	homeDir, err := os.UserHomeDir()
	usrStartup := filepath.Join(homeDir, "AppData", "Roaming",
		"Microsoft", "Windows", "Start Menu", "Programs", "Startup")

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)

	dir := ".winKidsTimer"
	cfgPath := filepath.Join(homeDir, dir, "config")
	dbPath := filepath.Join(homeDir, dir, "db")

	cfg := config{}

	pCfg, _ := pudge.Open(cfgPath, pudge.DefaultConfig)
	defer pCfg.Close()

	if err = pCfg.Get("cfg", &cfg); err != nil {
		log.Println(err)
		pCfg.Set("cfg", defaultConfig)
		err = nil
	}
	if cfg.Msg == "" {
		pCfg.Set("cfg", defaultConfig)
	}

	_, err = os.Stat(filepath.Join(usrStartup, processName))

	switch {
	case usrStartup == exPath:
		fmt.Println("запуск из автозагрузки")
		run(cfg, dbPath)
	case err == nil:
		edit(cfg, pCfg, usrStartup)
	default:
		setup(cfg, pCfg, usrStartup)
	}
}

type logs struct {
	Ts int64
}

func run(cfg config, dbPath string) {
	ts := time.Now()
	db, _ := pudge.Open(dbPath, pudge.DefaultConfig)

getKeys:
	keys, err := db.Keys(nil, 1, 0, false)
	if err != nil {
		log.Println(err)
	}

	log.Println(len(keys), cfg.WaitingSec, cfg.WaitingSecDeny, cfg.MaxPlay, cfg.PlayInterval)
	if len(keys) == 0 {
		db.Set(ts.Unix(), ts.Unix())
		goto getKeys
	}
	hh, mm, hh2, mm2 := cfg.Timers[0][0], cfg.Timers[0][1], cfg.Timers[0][2], cfg.Timers[0][3]

	low := time.Date(ts.Year(), ts.Month(), ts.Day(), hh, mm, 0, 0, ts.Location())
	high := time.Date(ts.Year(), ts.Month(), ts.Day(), hh2, mm2, 0, 0, ts.Location())

	if ts.Unix() > low.Unix() && ts.Unix() < high.Unix() {
		var last int64
		if err := db.Get(keys[0], &last); err != nil {
			log.Println(err)
		}
		maxPlay := int64(cfg.MaxPlay)
		playInterval := int64(cfg.PlayInterval)
		now := ts.Unix()
		switch {
		case (now - last) > (maxPlay + playInterval):
			db.Set(ts.Unix(), ts.Unix())
			goto getKeys
		case (now - last) < maxPlay:
			dur := time.Duration(maxPlay-(now-last)) * time.Second
			log.Println("Осталось играть", dur)
			// time.Sleep(dur)
			d := int(maxPlay - (now - last))
			// shutDown(cfg.WaitingSec, cfg.Msg)
			shutDown(d, cfg.Msg)
		case (now - last) > maxPlay:
			log.Println("Надо выключаться")
			shutDown(cfg.WaitingSecDeny, cfg.MsgDeny)
		}
	} else {
		log.Println("не попал в доступный интервал")
		shutDown(cfg.WaitingSecDeny, cfg.MsgDeny)
	}

}

func shutDown(wait int, msg string) {
	msg = strings.ReplaceAll(msg, " ", "_")
	m := fmt.Sprintf("shutdown -s -t %d -c \"%v\"", wait, msg)

	if err := exec.Command("cmd", "/C", m).Run(); err != nil {
		fmt.Println("Failed to initiate shutdown:", err)
	}
}

func edit(cfg config, pCfg *pudge.Db, usrStartup string) {
	mode := 0
point:
	fmt.Println("\nПрограмма в автозагрузке. Что делаем?")
	fmt.Println("1. Редактируем настройки")
	fmt.Println("2. Удаляем из автозагрузки")
	fmt.Println("3. Отменяем запланированное отключение ПК")
	fmt.Println("4. Ничего не делаем")
	fmt.Fscan(os.Stdin, &mode)

	switch mode {
	case 1:
		setup(cfg, pCfg, usrStartup)
	case 2:
		taskkill()
		delFromStartup(usrStartup)
		setup(cfg, pCfg, usrStartup)
	case 3:
		shutDownAbort()
	case 4:
		fmt.Println("\nПока.")
	default:
		fmt.Println("\nПопробуем ещё раз...\n")
		goto point
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
	fmt.Println("Готово.\n")
}

func setup(cfg config, pCfg *pudge.Db, usrStartup string) {
	mode := 0
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
		fmt.Println("\nПопробуем ещё раз...\n")
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
			fmt.Println("\nПопробуем ещё раз...\n")
			goto editLabel
		}
	case 3:
		mode = 0
		goto dynMode
	case 4:
		return
	default:
		fmt.Println("\nПопробуем ещё раз...\n")
		goto saveLabel
	}
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
	return out.Close()
}

func (c *config) Print() {
	fmt.Println()
	switch c.DynamicMode {
	case 1:
		fmt.Println("Интервалы, когда ПК может работать")
		for _, v := range c.Timers {
			fmt.Printf("%s:%s - %s:%s\n", getTime(v[0]), getTime(v[1]), getTime(v[2]), getTime(v[3]))
		}
	case 2:
		fmt.Print("Интервал, когда ПК может работать: ")
		for _, v := range c.Timers {
			fmt.Printf("%s:%s - %s:%s\n", getTime(v[0]), getTime(v[1]), getTime(v[2]), getTime(v[3]))
			break
		}
		fmt.Printf("Сколько доступно времени с момента включения: %d мин.\n", c.MaxPlay/60)
		fmt.Printf("Минимальный интервал между включениями: %d мин.\n", c.PlayInterval/60)
	}
}

func getTime(i int) string {
	if i < 10 {
		return "0" + getString(i)
	} else {
		return getString(i)
	}
}
