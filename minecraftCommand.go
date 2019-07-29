package main

import (
	"database/sql"
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
	"time"
)

func handleMinecraftCommand(s *discordgo.Session, message *discordgo.MessageCreate) {
	member, err := s.GuildMember(message.GuildID, message.Author.ID)
	if err != nil {
		log.Println("Błąd pobierania twórcy wiadomości!\n" + err.Error())
		return
	}
	if !hasRole(member, Config.PermittedRoleName, message.GuildID) {
		msg, err := s.ChannelMessageSend(message.ChannelID, Locale.NoPermission)
		if err == nil {
			time.Sleep(20 * time.Second)
			_ = s.ChannelMessageDelete(msg.ChannelID, msg.ID)
		}
		return
	}
	// dzielimy wiadomość po spacjach dla wygody
	args := strings.Split(message.Content, " ")
	if len(args) < 2 {
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.MinecraftInstruction)
		return
	}
	log.Println("Sprawdzam poprawność nicku " + args[1])
	if !validateMinecraftNickname(args[1]) {
		log.Println("Stwierdzam niepoprawność nicku " + args[1])
		msg, err := s.ChannelMessageSend(message.ChannelID, Locale.MinecraftIncorrectNickname)
		if err == nil {
			time.Sleep(20 * time.Second)
			_ = s.ChannelMessageDelete(msg.ChannelID, msg.ID)
		}
		return
	}
	log.Println("Stwierdzam poprawność nicku " + args[1])
	state := linkUserMinecraftNickname(message.Author.ID, args[1])
	if state == ERROR {
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.DatabaseError)
	} else if state == UPDATED {
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.MinecraftUpdatedNickname)
	} else if state == INSERTED {
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.MinecraftInsertedNickname)
	}

}

func linkUserMinecraftNickname(discordID, minecraftNickname string) State {
	var linkedUser LinkedUsers
	// sprawdzamy, czy takie id discorda jest już powiązane, unikamy duplikatów ,aktualizujemy.
	err := DbMap.SelectOne(&linkedUser, "SELECT * FROM LinkedUsers WHERE discord_id=?", discordID)
	// jeżeli nie ma wpisu z takim discord id...
	if err == sql.ErrNoRows {
		linkedUser.DiscordID = discordID
		linkedUser.MinecraftNickname.String = minecraftNickname
		linkedUser.MinecraftNickname.Valid = true
		err = DbMap.Insert(&linkedUser)
		if err != nil {
			log.Println("Błąd połączenia z bazą danych!\n" + err.Error())
			return ERROR
		}
		return INSERTED
	}
	if err != nil {
		log.Println("Błąd połączenia z bazą danych!\n" + err.Error())
		return ERROR
	}
	//a jeżeli takowy wpis jest
	linkedUser.MinecraftNickname.String = minecraftNickname
	linkedUser.MinecraftNickname.Valid = true
	_, err = DbMap.Update(&linkedUser)
	if err != nil {
		log.Println("Błąd połączenia z bazą danych!\n" + err.Error())
		return ERROR
	}
	return UPDATED
}

func validateMinecraftNickname(nick string) bool {
	if len(nick) < 3 || len(nick) > 16 {
		return false
	}
	for _, char := range nick {
		// wszystkie dopuszczone znaki w nicku w minecraft
		if !((char <= 'z' && char >= 'a') || (char <= 'Z' && char >= 'A') || (char <= '9' && char >= '0') || char == '_') {
			return false
		}
	}
	return true
}
