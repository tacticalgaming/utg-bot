package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"time"
)

type Config struct {
	Admins        []string `yaml:"admins"`          // Can do everything
	ServerAdmins  []string `yaml:"server_admins"`   // Can update/restart server
	ModAdmins     []string `yaml:"mod_admins"`      // Can update mods
	Uploaders     []string `yaml:"uploaders"`       // Can upload missions
	Directory     string   `yaml:"directory"`       // Home directory of arma user
	ServerDir     string   `yaml:"server_dir"`      // Path to game server directory
	ScriptsDir    string   `yaml:"scripts_dir"`     // Path to UTG scripts directory
	SteamCmd      string   `yaml:"steamcmd"`        // Path to steamcmd
	SteamUsername string   `yaml:"steam_user"`      // Steam username
	SteamPasswore string   `yaml:"steam_pass"`      // Steam password
	Uid           int      `yaml:"uid"`             // UID of arma user
	Gid           int      `yaml:"gid"`             // GID of arma user
	DiscordToken  string   `yaml:"discord_token"`   // Discord-bot token
	BotChannel    string   `yaml:"discord_channel"` // Channel bot spams to
}

func readConfig(path string) (*Config, error) {
	log.Infof("Reading configuration from %s", path)
	c := new(Config)

	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Config not found: %s", err.Error())
	}

	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return nil, fmt.Errorf("Couldn't parse yaml: %s", err.Error())
	}

	return c, nil
}

func main() {
	var err error
	log.Infof("Starting UTG Bot")
	config, err := readConfig("/etc/utg-bot.yaml")
	if err != nil {
		log.Errorf("Failed to read configuration: %s", err.Error())
		return
	}
	log.Infof("Reading mods table from %s", config.Directory+"/mods_table")

	mods := new(Mods)
	err = mods.Init(config)
	if err != nil {
		mods = nil
		log.Errorf("Failed to initialize mods: %s", err.Error())
	}

	missions := new(Missions)
	err = missions.Init(config)
	if err != nil {
		missions = nil
		log.Errorf("Failed to initialize missions: %s", err.Error())
	}

	server := new(Server)
	err = server.Init(config)
	if err != nil {
		server = nil
		log.Errorf("Failed to initialize game server: %s", err.Error())
	}

	b := new(Bot)
	err = b.Init(mods, missions, server, config.DiscordToken, config.BotChannel)
	if err != nil {
		log.Fatalf("Failed to init discord bot: %s", err.Error())
	}
	for {
		time.Sleep(time.Millisecond * 100)
	}
}
