package main

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

var raidMessagesId []string
var raidMessagesMutex sync.Mutex
var raidNotifierChannel = make(chan bool, 1)

func startRaid(raid Raid) {
	guild, err := session.Guild(Config.ServerId)
	if err != nil {
		log.Println("Błąd pobierania gildii.", err)
		return
	}
	for _, ch := range guild.Channels {
		role, err := getRoleID(Config.ServerId, "@everyone")
		if err != nil {
			log.Println("Błąd pobierania roli.", err)
			return
		}
		channel, err := session.Channel(ch.ID)
		if err != nil {
			log.Println("Błąd pobierania kanału.", err)
			return
		}
		var perms ChannelPermissions
		perms.RaidId = raid.Id
		perms.ChannelId = channel.ID
		found := false
		for _, permissionOverwrite := range channel.PermissionOverwrites {

			if permissionOverwrite.Type == "role" {
				if permissionOverwrite.ID == role {
					perms.EveryonePermissionsAllowed = permissionOverwrite.Allow
					perms.EveryonePermissionsDenied = permissionOverwrite.Deny
					found = true
					break
				}
			}
		}
		if !found {
			perms.EveryonePermissionsDenied = 0
			perms.EveryonePermissionsAllowed = 0
		}
		err = DbMap.Insert(&perms)
		if err != nil {
			log.Println("Unable to insert permissions!")
		}
	}
}

func endRaid(raid Raid) {
	raid.EndTime.Time = time.Now()
	raid.EndTime.Valid = true

	var permissions []ChannelPermissions
	_, err := DbMap.Select(&permissions, "SELECT * FROM ChannelsPermissions WHERE raid_id=?", raid.Id)
	if err != nil {
		log.Println("Niepowodzenie pobierania danych z bazy ", err)
		return
	}
	// eeee
	_, err = DbMap.Update(&raid)
	if err != nil {
		log.Println("Niepowodzenie aktualizacji danych w bazie ", err)
		return
	}
}

func handleRaidCommand(s *discordgo.Session, message *discordgo.MessageCreate) {
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
	var minutes int
	if len(args) == 2 {
		minutes, err = strconv.Atoi(args[1])
		if err != nil {
			log.Println("Błąd argumentu dni \"" + args[1] + "\" " + err.Error())
			msg, err := s.ChannelMessageSend(message.ChannelID, Locale.ColorIncorrectDaysCount)
			if err == nil {
				time.Sleep(20 * time.Second)
				_ = s.ChannelMessageDelete(msg.ChannelID, msg.ID)
			}
			return
		}
	} else {
		minutes = -1
	}
	var raid Raid
	raid.IssuerId = message.Author.ID
	var msg *discordgo.Message
	if minutes > 0 {
		msg, err = s.ChannelMessageSend(message.ChannelID, strings.ReplaceAll(Locale.RaidConfirmationTimed, "{MINUTES}", args[1]))
	} else {
		msg, err = s.ChannelMessageSend(message.ChannelID, Locale.RaidConfirmation)
	}
	if err != nil {
		log.Println("Niepowodzenie wysyłania wiadomości ", err)
		return
	}
	raid.MessageId = msg.ID
	raid.ChannelId = msg.ChannelID
	raid.Duration = minutes
	err = s.MessageReactionAdd(msg.ChannelID, msg.ID, "✅")
	if err != nil {
		log.Println("Nieudane dodawanie reakcji ", err)
		return
	}
	raidMessagesMutex.Lock()
	raidMessagesId = append(raidMessagesId, raid.MessageId)
	raidMessagesMutex.Unlock()
	select {
	case <-raidNotifierChannel:
		break
	case <-time.After(5 * time.Minute):
		err = s.MessageReactionsRemoveAll(msg.ChannelID, msg.ID)
		if err != nil {
			log.Println("niepowodzenie usuwania reakcji ", err)
			return
		}
		_, err = s.ChannelMessageEdit(msg.ChannelID, msg.ID, Locale.RaidRefused)
		if err != nil {
			log.Println("niepowodzenie edytowania wiadomości ", err)
			return
		}
	}
	raid.StartTime = time.Now()

	err = DbMap.Insert(&raid)
	if err != nil {
		log.Println("Niepowodzenie wprowadzania danych do bazy danych ", err)
		return
	}

	startRaid(raid)
	if minutes > 0 {
		time.Sleep(time.Duration(minutes) * time.Minute)
		endRaid(raid)
	}

}
