package main

import (
	"database/sql"
	"github.com/bopke/MultisquadDiscordBot/database"
	"github.com/bwmarrin/discordgo"
	"log"
	"time"
)

func handleUnvipCommand(s *discordgo.Session, message *discordgo.MessageCreate) {
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
	if message.Mentions != nil && len(message.Mentions) == 0 {
		msg, err := s.ChannelMessageSend(message.ChannelID, Locale.VipIncorrectUser)
		if err == nil {
			time.Sleep(20 * time.Second)
			_ = s.ChannelMessageDelete(msg.ChannelID, msg.ID)
		}
		return
	}
	var linkedUser database.LinkedUsers
	err = database.DbMap.SelectOne(&linkedUser, "SELECT * FROM LinkedUsers WHERE discord_id = ?", message.Mentions[0].ID)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Błąd połączenia z bazą danych!\n" + err.Error())
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.DatabaseError)
		return
	}
	if err == sql.ErrNoRows {
		msg, err := s.ChannelMessageSend(message.ChannelID, "Ten użytkownik nigdy nie był vipem.")
		if err == nil {
			time.Sleep(20 * time.Second)
			_ = s.ChannelMessageDelete(msg.ChannelID, msg.ID)
		}
		return
	}
	linkedUser.Valid = false
	linkedUser.ExpirationDate = time.Now()
	linkedUser.NotifiedExpiration = true
	_, err = database.DbMap.Update(&linkedUser)
	role, err := getRoleID(message.GuildID, "VIP")
	err = s.GuildMemberRoleRemove(message.GuildID, message.Mentions[0].ID, role)
	if err != nil {
		log.Println("Błąd odbierania rangi " + err.Error())
		msg, err := s.ChannelMessageSend(message.ChannelID, Locale.UnexpectedApiError)
		if err == nil {
			time.Sleep(20 * time.Second)
			_ = s.ChannelMessageDelete(msg.ChannelID, msg.ID)
		}
		return
	}
	log.Println("Status VIP użytkownika " + message.Mentions[0].Username + "#" + message.Mentions[0].Discriminator + " został zakończony")
}
