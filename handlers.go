package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os/exec"
	"time"
)

var lastRestart time.Time = time.Unix(0, 0)

func handleRestart(authorID string) (string, error) {
	log.Infof("Restart requested by %s", authorID)
	if time.Since(lastRestart) < time.Minute*5 {
		return "Сервер можно перезапускать только раз в 5 минут", fmt.Errorf("Restart time limit")
	}

	cmd := exec.Command("/bin/systemctl", "restart", "utg")
	err := cmd.Run()
	if err != nil {
		return "Не получилось перезапустить: " + err.Error(), err
	}

	lastRestart = time.Now()

	return "Сервер был перезапущен", nil
}

func handleHelp() (string, error) {
	text := "Я умею:\n"
	text += "!mods - управлять модами\n"
	text += "!server - управлять сервером\n"
	text += "Ты всегда можешь узнать подробности написав !server help, !mods help и т.д.\n"
	text += "Скоро я буду уметь больше ;)\n"
	return text, nil
}

func handleServer(authorID string, param []string) (string, error) {
	return getServerHelp(), nil
}
