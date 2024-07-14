package main

import (
	jsoniter "github.com/json-iterator/go"
	"log"
	"os"
)

var defaultConfig = &config{
	From:         "07h",
	To:           "21h",
	WaitingSec:   30,
	MaxPlay:      "1h",
	PlayInterval: "2h",
	Msg:          "Я хочу спать!",
	MsgDeny:      "Я не выспался. Посплю ещё",
}

type config struct {
	From         string
	To           string
	WaitingSec   int
	MaxPlay      string
	PlayInterval string
	Msg          string
	MsgDeny      string
}

func (c *config) init() {
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		log.Printf("Конфиг не обнаружен. %v\nСоздан новый с настройками по-умолчанию", err)
		cfg = defaultConfig
		c.write()
		return
	}
	if err := jsoniter.Unmarshal(data, cfg); err != nil {
		log.Printf("Ошибка при разборе конфига. %v\nСоздан новый с настройками по-умолчанию", err)
		cfg = defaultConfig
		c.write()
	}
}
func (c *config) write() {
	data, err := jsoniter.Marshal(cfg)
	if err != nil {
		log.Println(err)
		return
	}
	if err := os.WriteFile(cfgPath, data, os.ModePerm); err != nil {
		log.Println(err)
	}
}
