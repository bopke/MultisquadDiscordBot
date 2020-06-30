package main

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
	"time"
)

func handleAnnounceCommand(s *discordgo.Session, message *discordgo.MessageCreate) {
	_ = s.ChannelMessageDelete(message.ChannelID, message.ID)
	member, err := s.GuildMember(message.GuildID, message.Author.ID)
	if err != nil {
		log.Println("Błąd pobierania twórcy wiadomości!\n" + err.Error())
		return
	}
	if !hasPermission(member, message.GuildID, discordgo.PermissionAdministrator) {
		msg, err := s.ChannelMessageSend(message.ChannelID, Locale.NoAdminPermission)
		if err == nil {
			time.Sleep(20 * time.Second)
			_ = s.ChannelMessageDelete(msg.ChannelID, msg.ID)
		}
		return
	}
	// dzielimy wiadomość po spacjach dla wygody
	args := strings.Split(message.Content, " ")
	_, _ = s.ChannelMessageSend(message.ChannelID, strings.Join(args[1:], " "))
}
