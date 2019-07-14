package main

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
)

// funkcja ta przyjmuje każdą wiadomość która zostanie wysłana na kanałach, które widzi bot i analizuje ją.
func OnMessageCreate(s *discordgo.Session, message *discordgo.MessageCreate) {
	//jeżeli wiadomość jest na serwerze innym niż nasz oczekiwany to wywalać z tymi komendami.
	if message.GuildID != Config.ServerId {
		return
	}
	// jeżeli wiadomość nie zaczyna się od naszej komendy to nie analizujemy dalej
	if !strings.HasPrefix(message.Content, Config.CommandName) {
		return
	}
	member, err := s.GuildMember(message.GuildID, message.Author.ID)
	if err != nil {
		log.Println("Błąd pobierania twórcy wiadomości!\n" + err.Error())
		return
	}
	if !hasRole(member, Config.PermittedRoleName, message.GuildID) {
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.NoPermission)
		return
	}
	log.Println(message.Author.Username + "#" + message.Author.Discriminator + " wykonał polecenie: " + message.Content)
	// dzielimy wiadomość po spacjach dla wygody
	args := strings.Split(message.Content, " ")
	if len(args) < 2 {
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.InvalidProfileLink)
		return
	}
	// link do profilu może zacząć się od https:// ...
	if strings.HasPrefix(args[1], "https://") {
		args[1] = args[1][8:]
	} else
	// lub od http://...
	if strings.HasPrefix(args[1], "http://") {
		args[1] = args[1][7:]
	}
	//albo po prostu od razu od adresu do strony steama. Pozwólmy ludziom wklejać różne warianty.
	if !strings.HasPrefix(args[1], "steamcommunity.com/id/") {
		if strings.HasPrefix(args[1], "steamcommunity.com/profiles/") {
			// interesuje nas tylko to, co jest po /profiles/
			args[1] = args[1][28:]
			err := validateSteamId(args[1])
			if err == nil {
				state := linkUserSteamID(message.Author.ID, args[1])
				if state == ERROR {
					_, _ = s.ChannelMessageSend(message.ChannelID, Locale.DatabaseError)
				} else if state == UPDATED {
					_, _ = s.ChannelMessageSend(message.ChannelID, Locale.SteamIdUpdated)
				} else if state == INSERTED {
					_, _ = s.ChannelMessageSend(message.ChannelID, Locale.SteamIdInserted)
				}
				return
			} else {
				_, _ = s.ChannelMessageSend(message.ChannelID, Locale.UnexpectedError)
				return
			}
		}
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.InvalidProfileLink)
		return
	}
	//interesuje nas tylko to, co jest po id/
	args[1] = args[1][22:]
	// i na podstawie tego możemy pobrać steamID
	steamId, err := getSteamIdForProfileId(args[1])
	if err != nil {
		// ale ktoś może podać zły link i taki profil nie istnieje!
		if err == NoSuchProfileError {
			_, _ = s.ChannelMessageSend(message.ChannelID, Locale.InvalidProfileId)
			return
		}
		// może też wystąpić jakiś nieoczekiwany błąd..
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.UnexpectedError)
		log.Println("Błąd przy odczycie danych z API!\n" + err.Error())
		return
	}
	state := linkUserSteamID(message.Author.ID, steamId)
	if state == ERROR {
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.DatabaseError)
	} else if state == UPDATED {
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.SteamIdUpdated)
	} else if state == INSERTED {
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.SteamIdInserted)
	}
}
