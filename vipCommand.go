package main

import (
	"database/sql"
	"github.com/bwmarrin/discordgo"
	"log"
	"strconv"
	"strings"
	"time"
)

func handleVipCommand(s *discordgo.Session, message *discordgo.MessageCreate) {
	_ = s.ChannelMessageDelete(message.ChannelID, message.ID)
	member, err := s.GuildMember(message.GuildID, message.Author.ID)
	if err != nil {
		log.Println("Błąd pobierania twórcy wiadomości!\n" + err.Error())
		return
	}
	if !hasPermission(member, message.GuildID, discordgo.PermissionAdministrator) {
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.NoAdminPermission)
		return
	}
	// dzielimy wiadomość po spacjach dla wygody
	args := strings.Split(message.Content, " ")
	var length int
	if message.Mentions != nil && len(message.Mentions) == 0 {
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.VipIncorrectUser)
		return
	}
	if len(args) >= 3 {
		length, err = strconv.Atoi(args[2])
		if err != nil {
			log.Println("Błąd argumentu dni \"" + args[2] + "\" " + err.Error())
			_, _ = s.ChannelMessageSend(message.ChannelID, Locale.VipIncorrectDaysCount)
			return
		}
	} else {
		length = Config.VipDefaultLength
	}
	var linkedUser LinkedUsers
	err = DbMap.SelectOne(&linkedUser, "SELECT * FROM LinkedUsers WHERE discord_id = ?", message.Mentions[0].ID)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Błąd połączenia z bazą danych!\n" + err.Error())
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.DatabaseError)
		return
	}
	if err == sql.ErrNoRows {
		linkedUser.DiscordID = message.Mentions[0].ID
		linkedUser.Valid = true
		linkedUser.ExpirationDate = time.Now().Add((time.Hour * 24) * time.Duration(length))
		linkedUser.NotifiedExpiration = false
		err = DbMap.Insert(&linkedUser)
		if err != nil {
			log.Println("Błąd połączenia z bazą danych!\n" + err.Error())
			_, _ = s.ChannelMessageSend(message.ChannelID, Locale.DatabaseError)
			return
		}
		role, err := getRoleID(message.GuildID, Config.PermittedRoleName)
		if err != nil {
			log.Println("Nie udalo sie pobrac roli vip " + err.Error())
			_, _ = s.ChannelMessageSend(message.ChannelID, Locale.UnexpectedApiError)
			return
		}
		err = s.GuildMemberRoleAdd(message.GuildID, message.Mentions[0].ID, role)
		if err != nil {
			log.Println("Błąd nadawania rangi " + err.Error())
			_, _ = s.ChannelMessageSend(message.ChannelID, Locale.UnexpectedApiError)
			return
		}
		log.Println("Status VIP użytkownika " + message.Mentions[0].Username + "#" + message.Mentions[0].Discriminator + " został utworzony na " + strconv.Itoa(length) + " dni")
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.VipInserted)
		_, _ = s.ChannelMessageSend(Config.AnnouncementChannelId, strings.Replace(Locale.VipAnnouncementInserted, "{MENTION}", message.Mentions[0].Mention(), -1))
		return
	}
	linkedUser.Valid = true
	if linkedUser.ExpirationDate.Before(time.Now()) {
		linkedUser.ExpirationDate = time.Now().Add((time.Hour * 24) * time.Duration(length))
	} else {
		linkedUser.ExpirationDate = linkedUser.ExpirationDate.Add((time.Hour * 24) * time.Duration(length))
	}
	linkedUser.NotifiedExpiration = false
	_, err = DbMap.Update(&linkedUser)
	if err != nil {
		log.Println("Błąd połączenia z bazą danych!\n" + err.Error())
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.DatabaseError)
		return
	}
	role, err := getRoleID(message.GuildID, Config.PermittedRoleName)
	if err != nil {
		log.Println("Nie udalo sie pobrac roli vip " + err.Error())
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.UnexpectedApiError)
		return
	}
	err = s.GuildMemberRoleAdd(message.GuildID, message.Mentions[0].ID, role)
	if err != nil {
		log.Println("Błąd nadawania rangi " + err.Error())
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.UnexpectedApiError)
		return
	}
	log.Println("Status VIP użytkownika " + message.Mentions[0].Username + "#" + message.Mentions[0].Discriminator + " został utworzony na " + strconv.Itoa(length) + " dni")
	_, _ = s.ChannelMessageSend(message.ChannelID, Locale.VipUpdated)
	_, _ = s.ChannelMessageSend(Config.AnnouncementChannelId, strings.Replace(Locale.VipAnnouncementUpdated, "{MENTION}", message.Mentions[0].Mention(), -1))
}
