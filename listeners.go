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
	// jeżeli wiadomość zaczyna się od naszej komendy to analizujemy dalej
	if strings.HasPrefix(message.Content, Config.SteamCommandName) {
		log.Println(message.Author.Username + "#" + message.Author.Discriminator + " wykonał polecenie: " + message.Content)
		handleSteamCommand(s, message)
		return
	}
	if strings.HasPrefix(message.Content, Config.StatusCommandName) {
		log.Println(message.Author.Username + "#" + message.Author.Discriminator + " wykonał polecenie: " + message.Content)
		handleStatusCommand(s, message)
		return
	}
	if strings.HasPrefix(message.Content, Config.MinecraftCommandName) {
		log.Println(message.Author.Username + "#" + message.Author.Discriminator + " wykonał polecenie: " + message.Content)
		handleMinecraftCommand(s, message)
		return
	}
	if strings.HasPrefix(message.Content, Config.VipCommandName) {
		log.Println(message.Author.Username + "#" + message.Author.Discriminator + " wykonał polecenie: " + message.Content)
		handleVipCommand(s, message)
	}
}
