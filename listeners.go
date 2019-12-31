package main

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"math/rand"
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

func tickRatelimitCounter(msg *discordgo.Message, duration int) {
	for ; duration > 0; duration-- {
		start := time.Now()
		_, _ = session.ChannelMessageEdit(msg.ChannelID, msg.ID, strings.Replace(Locale.RateLimitWait, "{SECONDS}", strconv.Itoa(duration), -1))
		time.Sleep(start.Add(time.Second).Sub(time.Now()))
	}
	_ = session.ChannelMessageDelete(msg.ChannelID, msg.ID)
}

func isTooEarlyToExecute(command string, message *discordgo.MessageCreate) bool {

	if rateLimits[command][message.Author.ID].Add(10 * time.Second).After(time.Now()) {
		timeDifference := rateLimits[command][message.Author.ID].Add(10 * time.Second).Sub(time.Now())
		timeDiff := int(timeDifference.Seconds()) + 1
		msg, err := session.ChannelMessageSend(message.ChannelID, strings.Replace(Locale.RateLimitWait, "{SECONDS}", strconv.Itoa(timeDiff), -1))
		if err == nil {
			go tickRatelimitCounter(msg, timeDiff-1)
		}
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

func OnGuildMemberUpdate(s *discordgo.Session, e *discordgo.GuildMemberUpdate) {
	if !Config.ChangeBotNicknames && e.User.Bot {
		return
	}
	fixNickname(e.Member)
}

func OnMessageReactionAdd(s *discordgo.Session, reaction *discordgo.MessageReactionAdd) {
	if reaction.MessageID != Config.RulesMessageId {
		return
	}
	if reaction.Emoji.Name != Config.RulesAgreementEmojiName {
		return
	}
	if len(Config.VerifiedRolesIds) == 0 {
		return
	}
	member, err := s.GuildMember(reaction.GuildID, reaction.UserID)
	if err != nil {
		log.Println("OnMessageReactionAdd Unable to get member! ", err)
		return
	}
	for _, roleId := range member.Roles {
		if roleId == Config.VerifiedRolesIds[0] {
			return
		}
	}
	wasAbleToAddAllRoles := true
	for _, roleId := range Config.VerifiedRolesIds {
		err = s.GuildMemberRoleAdd(reaction.GuildID, reaction.UserID, roleId)
		if err != nil {
			wasAbleToAddAllRoles = false
		}
	}
	embed := &discordgo.MessageEmbed{
		Title:       "Witaj na Young Multi",
		Description: "Siemano " + member.Mention() + "!\n- Przestrzegaj <#320578596223844353>\n- Odbierz rangi na kanale <#581907929184075783>\n- Baw się dobrze!",
		Timestamp:   time.Now().Format(time.RFC3339),
		Color:       rand.Intn(0xFFFFFF),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://media1.tenor.com/images/b7ac38f04efc899c84ff9975132f4add/tenor.gif",
		},
	}
	content := member.Mention() + "<@&661313646659633162>"
	if !wasAbleToAddAllRoles {
		content += "\n Nie udało mi się nadać Ci wszystkich ról :worried:"
	}
	_, err = s.ChannelMessageSendComplex(Config.AnnouncementChannelId, &discordgo.MessageSend{
		Content: content,
		Embed:   embed,
	})
	if err != nil {
		log.Println("OnMessageReactionAdd Unable to send channel message! ", err)
		return
	}
}
