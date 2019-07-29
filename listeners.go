package main

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"strconv"
	"strings"
	"time"
)

var rateLimits map[string]map[string]time.Time

func InitRateLimits() {
	rateLimits = map[string]map[string]time.Time{}
	rateLimits[Config.SteamCommandName] = map[string]time.Time{}
	rateLimits[Config.StatusCommandName] = map[string]time.Time{}
	rateLimits[Config.MinecraftCommandName] = map[string]time.Time{}
}

func isTooEarlyToExecute(command string, message *discordgo.MessageCreate) bool {

	if rateLimits[command][message.Author.ID].Add(10 * time.Second).After(time.Now()) {
		timeDiff := rateLimits[command][message.Author.ID].Add(10 * time.Second).Sub(time.Now())
		_, _ = session.ChannelMessageSend(message.ChannelID, strings.Replace(Locale.RateLimitWait, "{SECONDS}", strconv.Itoa(int(timeDiff.Seconds())+1), -1))
		return true
	}
	rateLimits[command][message.Author.ID] = time.Now()
	return false
}

// funkcja ta przyjmuje każdą wiadomość która zostanie wysłana na kanałach, które widzi bot i analizuje ją.
func OnMessageCreate(s *discordgo.Session, message *discordgo.MessageCreate) {
	//jeżeli wiadomość jest na serwerze innym niż nasz oczekiwany to wywalać z tymi komendami.
	if message.GuildID != Config.ServerId {
		return
	}
	// jeżeli wiadomość zaczyna się od naszej komendy to analizujemy dalej
	if strings.HasPrefix(message.Content, Config.SteamCommandName) {
		if isTooEarlyToExecute(Config.SteamCommandName, message) {
			return
		}
		log.Println(message.Author.Username + "#" + message.Author.Discriminator + " wykonał polecenie: " + message.Content)
		handleSteamCommand(s, message)
		return
	}
	if strings.HasPrefix(message.Content, Config.StatusCommandName) {
		if isTooEarlyToExecute(Config.StatusCommandName, message) {
			return
		}
		log.Println(message.Author.Username + "#" + message.Author.Discriminator + " wykonał polecenie: " + message.Content)
		handleStatusCommand(s, message)
		return
	}
	if strings.HasPrefix(message.Content, Config.MinecraftCommandName) {
		if isTooEarlyToExecute(Config.MinecraftCommandName, message) {
			return
		}
		log.Println(message.Author.Username + "#" + message.Author.Discriminator + " wykonał polecenie: " + message.Content)
		handleMinecraftCommand(s, message)
		return
	}
	if strings.HasPrefix(message.Content, Config.VipCommandName) {
		log.Println(message.Author.Username + "#" + message.Author.Discriminator + " wykonał polecenie: " + message.Content)
		handleVipCommand(s, message)
	}
}
