package main

import (
	"github.com/bopke/MultisquadDiscordBot/context"
	"github.com/bopke/MultisquadDiscordBot/util"
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
)

func handleAnnounceCommand(s *discordgo.Session, message *discordgo.MessageCreate) {
	_ = s.ChannelMessageDelete(message.ChannelID, message.ID)
	member, err := s.GuildMember(message.GuildID, message.Author.ID)
	if err != nil {
		log.Println("Błąd pobierania twórcy wiadomości!\n" + err.Error())
		return
	}
	if !util.HasPermission(&context.Context{
		Session:   s,
		Guild:     nil,
		Member:    member,
		Message:   message.Message,
		ChannelId: message.ChannelID,
		GuildId:   message.GuildID,
		UserId:    member.User.ID,
		MessageId: message.ID,
	}, discordgo.PermissionAdministrator) {
		return
	}
	// dzielimy wiadomość po spacjach dla wygody
	args := strings.Split(message.Content, " ")
	_, _ = s.ChannelMessageSend(message.ChannelID, strings.Join(args[1:], " "))
}
