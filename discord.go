package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"strings"
)

type Bot struct {
	TokenKey     string             // Token for bot (discord)
	Session      *discordgo.Session // Discord session
	ChannelID    string             // Discord channel ID
	LogChannelID string             // Log channel ID
	mods         *Mods
	missions     *Missions
	server       *Server
}

func (b *Bot) Init(mods *Mods, missions *Missions, server *Server, token, logChannel string) error {
	b.TokenKey = token
	b.LogChannelID = logChannel
	b.mods = mods
	b.missions = missions
	var err error
	b.Session, err = discordgo.New("Bot " + b.TokenKey)
	if err != nil {
		return fmt.Errorf("Failed to create discord session: %s", err)
	}
	b.Session.Token = "Bot " + b.TokenKey
	b.Session.State.User, err = b.Session.User("@me")
	if err != nil {
		return fmt.Errorf("Failed to retrieve user data: %s", err)
	}

	//log.Infof("Bot ID: %s [%s]", b.Session.State.User.ID, b.Session.State.User.Username)

	/*
		guilds, err := b.Session.UserGuilds(100, "", "")
		if err != nil {
			return fmt.Errorf("Failed to retrieve guilds")
		}

		guild, err := b.Session.Guild(guilds[0].ID)
		if err != nil {
			return fmt.Errorf("Failed to retrieve guild")
		}

		channels, err := b.Session.GuildChannels(guild.ID)
		if err != nil {
			return fmt.Errorf("Failed to retrieve channels")
		}
	*/

	b.Session.AddHandler(b.messageCreate)

	err = b.Session.Open()
	if err != nil {
		return fmt.Errorf("Failed to open a connection to Discord: %s", err)
	}

	return nil
}

func (b *Bot) messageCreate(s *discordgo.Session, msg *discordgo.MessageCreate) {

	parts := strings.Split(msg.Content, " ")
	if len(parts) == 0 {
		return
	}

	// Starts with an exclamation mark. Probably a command
	if len(parts[0]) > 0 && parts[0][0] == '!' {
		switch parts[0] {
		case "!restart":
			b.response(fmt.Sprintf("Перезапускаю сервер по просьбе %s", msg.Author.Mention()))
			response, err := handleRestart(msg.Author.ID)
			if err != nil {
				log.Infof("Restart request failed: %s", err.Error())
			}
			b.response(response)
		case "!mods":
			if b.mods != nil {
				response, err := b.mods.handle(msg.Author.ID, parts[1:])
				if err != nil {
					log.Infof("Restart request failed: %s", err.Error())
				}
				b.response(response)
			}
		case "!missions":
			if b.missions != nil {
				response, err := b.missions.handle(msg.Author.ID, parts[1:], msg.Attachments)
				if err != nil {
					log.Infof("Failed to handle missions request: %s", err.Error())
				}
				b.response(response)
			}
		case "!server":
			if b.server != nil {
				response, err := b.server.handle(msg.Author.ID, parts[1:])
				if err != nil {
					log.Infof("Restart request failed: %s", err.Error())
				}
				b.response(response)
			}
		case "!help":
			response, _ := handleHelp()
			b.response(response)
		}
	}
}

func (b *Bot) response(text string) {
	buffer := []string{}

	// Discord can receive message up to 2000 characters long.
	if len(text) > 1999 {
		lines := strings.Split(text, "\n")
		b := ""
		for _, line := range lines {
			if len(b)+len(line)+len("\n") > 1999 {
				buffer = append(buffer, b)
				b = line + "\n"
				continue
			}
			b += line + "\n"
		}
		buffer = append(buffer, b)
	} else {
		buffer = append(buffer, text)
	}

	for _, str := range buffer {
		_, err := b.Session.ChannelMessageSend(b.LogChannelID, str)
		if err != nil {
			log.Errorf("Failed to send message to %s: %s", b.LogChannelID, err.Error())
		}
	}
}
