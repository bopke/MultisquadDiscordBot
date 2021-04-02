package main

import (
	"github.com/bopke/MultisquadDiscordBot/commands"
	"github.com/bopke/MultisquadDiscordBot/config"
	"github.com/bopke/MultisquadDiscordBot/money"
	"github.com/bopke/MultisquadDiscordBot/nicks"
	"github.com/bwmarrin/discordgo"
	"log"
	"math/rand"
	"strings"
	"time"
)

func OnDMMessageReactionAdd(s *discordgo.Session, reaction *discordgo.MessageReactionAdd) {
	channel, err := s.Channel(reaction.ChannelID)
	if err != nil {
		log.Println("OnMessageCreate Unable to retrieve channel ", err)
		return
	}
	if channel.Type == discordgo.ChannelTypeDM {
		handleDMReactions(s, reaction, channel)
		return
	}
}

// funkcja ta przyjmuje każdą wiadomość która zostanie wysłana na kanałach, które widzi bot i analizuje ją.
func OnMessageCreate(s *discordgo.Session, message *discordgo.MessageCreate) {
	//	log.Println(message.Content)
	if s.State.User.ID != message.Author.ID && message.Author.Bot {
		return
	}
	addReactionToImage(message.Message)
	channel, err := s.Channel(message.ChannelID)
	if err != nil {
		log.Println("OnMessageCreate Unable to retrieve channel ", err)
		return
	}
	if channel.Type == discordgo.ChannelTypeDM {
		handleDMMessages(s, message, channel)
		return
	}

	//jeżeli wiadomość jest na serwerze innym niż nasz oczekiwany to wywalać z tymi komendami.
	if message.GuildID != config.GuildId {
		return
	}

	go money.HandleMessageMoneyCount(s, message)
	// jeżeli wiadomość zaczyna się od naszej komendy to analizujemy dalej
	if strings.HasPrefix(message.Content, "!steam") {
		log.Println(message.Author.Username + "#" + message.Author.Discriminator + " wykonał polecenie: " + message.Content)
		handleSteamCommand(s, message)
		return
	}
	if strings.HasPrefix(message.Content, "!status") {
		log.Println(message.Author.Username + "#" + message.Author.Discriminator + " wykonał polecenie: " + message.Content)
		handleStatusCommand(s, message)
		return
	}
	if strings.HasPrefix(message.Content, "!minecraft") {
		log.Println(message.Author.Username + "#" + message.Author.Discriminator + " wykonał polecenie: " + message.Content)
		handleMinecraftCommand(s, message)
		return
	}
	if strings.HasPrefix(message.Content, "!vips") {
		log.Println(message.Author.Username + "#" + message.Author.Discriminator + " wykonał polecenie: " + message.Content)
		handleVipsCommand(s, message)
		return
	}
	if strings.HasPrefix(message.Content, "!unvip") {
		log.Println(message.Author.Username + "#" + message.Author.Discriminator + " wykonał polecenie: " + message.Content)
		handleUnvipCommand(s, message)
		return
	}
	if strings.HasPrefix(message.Content, "!announce") {
		log.Println(message.Author.Username + "#" + message.Author.Discriminator + " wykonał polecenie: " + message.Content)
		handleAnnounceCommand(s, message)
		return
	}
	if strings.HasPrefix(message.Content, "!report") {
		log.Println(message.Author.Username + "#" + message.Author.Discriminator + " wykonał polecenie: " + message.Content)
		handleReportCommand(s, message)
		return
	}
	if (strings.HasPrefix(message.Content, "!monety") || strings.HasPrefix(message.Content, "!mon")) && (message.ChannelID == "597216492123324442" || message.ChannelID == "698639658732486846" || message.ChannelID == "771544460416778250" || message.ChannelID == "581950409094987838") {
		log.Println(message.Author.Username + "#" + message.Author.Discriminator + " wykonał polecenie: " + message.Content)
		handleMoneyCommand(s, message)
	}
	/*	if strings.HasPrefix(message.Content, Config.RaidCommandName) {
		log.Println(message.Author.Username + "#" + message.Author.Discriminator + " wykonał polecenie: " + message.Content)
		handleRaidCommand(s, message)
		return
	}*/
	commands.Listener(s, message)
}

func OnGuildMemberUpdate(s *discordgo.Session, e *discordgo.GuildMemberUpdate) {
	if e.User.Bot {
		return
	}
	nicks.FixNickname(s, e.Member)
}

const (
	RulesAgreementEmojiName = "tak"
	RulesMessageId          = "795633406192648242"
)

func OnMessageReactionAdd(s *discordgo.Session, reaction *discordgo.MessageReactionAdd) {
	if reaction.MessageID != RulesMessageId {
		return
	}
	if reaction.Emoji.Name != RulesAgreementEmojiName {
		return
	}
	if len(config.VerifiedRolesIds) == 0 {
		return
	}
	member, err := s.GuildMember(reaction.GuildID, reaction.UserID)
	if err != nil {
		log.Println("OnMessageReactionAdd Unable to get member! ", err)
		return
	}
	for _, roleId := range member.Roles {
		if roleId == config.VerifiedRolesIds[0] {
			return
		}
	}
	wasAbleToAddAllRoles := true
	for _, roleId := range config.VerifiedRolesIds {
		err = s.GuildMemberRoleAdd(reaction.GuildID, reaction.UserID, roleId)
		if err != nil {
			wasAbleToAddAllRoles = false
		}
	}
	embed := &discordgo.MessageEmbed{
		Title:       "Witaj na Young Multi",
		Description: "Siemano " + member.Mention() + "!\n- Przestrzegaj <#320578596223844353>\n- Odbierz rangi na kanale <#662764876355207228>\n- Baw się dobrze!",
		Timestamp:   time.Now().Format(time.RFC3339),
		Color:       rand.Intn(0xFFFFFF),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://media1.tenor.com/images/b7ac38f04efc899c84ff9975132f4add/tenor.gif",
		},
	}
	content := member.Mention()
	if !wasAbleToAddAllRoles {
		content += "\n Nie udało mi się nadać Ci wszystkich ról :worried:"
	}
	_, err = s.ChannelMessageSendComplex(config.AnnouncementChannelId, &discordgo.MessageSend{
		Content: content,
		Embed:   embed,
	})
	if err != nil {
		log.Println("OnMessageReactionAdd Unable to send channel message! ", err)
		return
	}
}

func OnGuildMemberAdd(s *discordgo.Session, e *discordgo.GuildMemberAdd) {
	if e.User.Bot {
		return
	}
	nicks.FixNickname(s, e.Member)
}
