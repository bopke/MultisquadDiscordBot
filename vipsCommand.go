package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"time"
)

func handleVipsCommand(s *discordgo.Session, message *discordgo.MessageCreate) {
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
	var actualVips []LinkedUsers
	_, err = DbMap.Select(&actualVips, "SELECT discord_id,expiration_date FROM LinkedUsers WHERE expiration_date >= NOW() ORDER BY expiration_date")
	if err != nil {
		log.Println("Błąd pobierania aktualnych vipów.", err)
		return
	}
	log.Println("Pobrałem informacje o", len(actualVips), "vipach")
	content := ""
	for i, actualVip := range actualVips {
		content += "<@" + actualVip.DiscordID + "> - " + fmt.Sprintf("%4d-%2d-%2d %2d:%2d\n", actualVip.ExpirationDate.Year(), actualVip.ExpirationDate.Month(), actualVip.ExpirationDate.Day(), actualVip.ExpirationDate.Hour(), actualVip.ExpirationDate.Minute())
		if i > 0 && i%20 == 0 {
			embed := discordgo.MessageEmbed{
				Title:       "Aktualne VIPy",
				Description: content,
				Timestamp:   time.Now().Format(time.RFC3339),
			}
			_, err = s.ChannelMessageSendEmbed(message.ChannelID, &embed)
			if err != nil {
				log.Println("Błąd wysyłania embeda.", err)
			}
		}
	}
	if len(content) != 0 {
		embed := discordgo.MessageEmbed{
			Title:       "Aktualne VIPy",
			Description: content,
			Timestamp:   time.Now().Format(time.RFC3339),
		}
		_, err = s.ChannelMessageSendEmbed(message.ChannelID, &embed)
		if err != nil {
			log.Println("Błąd wysyłania embeda.", err)
		}
	}
	log.Println("Skończyłem")
}
