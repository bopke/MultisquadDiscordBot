package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"time"
)

func handleShopCommand(s *discordgo.Session, message *discordgo.MessageCreate) {
	embed := &discordgo.MessageEmbed{
		URL:         "",
		Type:        "",
		Title:       "",
		Description: "Kup przedmiot za pomocą komendy `!kup`",
		Timestamp:   time.Now().Format(time.RFC3339),
		Color:       0x00BB00,
		Footer:      nil,
		Image:       nil,
		Thumbnail:   nil,
		Video:       nil,
		Provider:    nil,
		Author: &discordgo.MessageEmbedAuthor{
			Name:         message.Author.Username + "#" + message.Author.Discriminator,
			IconURL:      message.Author.AvatarURL("128"),
			ProxyIconURL: "",
		},
		Fields: []*discordgo.MessageEmbedField{
			{Name: "1. Nick 500 <a:moneta:613020692346175628>",
				Value:  "Zmiana nicku. Kup poleceniem `!kup nick <nowy nick>`, np `!kup nick Young Multi`.",
				Inline: false,
			},
			{Name: "2. Kolor 15.000 <a:moneta:613020692346175628>",
				Value:  "Zmiana koloru nicku. Kup poleceniem `!kup kolor <nazwa koloru>`, np `!kup kolor pomaranczowy`.\nDostępne kolory: pomarańczowy, jakieś jeszcze idk",
				Inline: false,
			},
			{Name: "3. Mysterybox 5.000 <a:moneta:613020692346175628>",
				Value:  "Losowanie itemka (czymkolwiek to jest). Kup poleceniem `!kup mysterybox`.",
				Inline: false,
			},
			{Name: "4. VIP 65.000 <a:moneta:613020692346175628>",
				Value:  "VIP na 30 dni. Kup poleceniem `!kup vip`.",
				Inline: false,
			},
			{Name: "5. Nitro 90.000 <a:moneta:613020692346175628>",
				Value:  "Nitro na 30 dni. Wysyłane ręcznie przez administratora. Kup poleceniem `!kup nitro`.",
				Inline: false,
			},
			{Name: "6. Flexer 100.000 <a:moneta:613020692346175628>",
				Value:  "Ranga Flexer. Kup poleceniem `!kup flexer`.",
				Inline: false,
			},
		},
	}
	_, _ = s.ChannelMessageSendEmbed(message.ChannelID, embed)
}

func handleBuyCommand(s *discordgo.Session, message *discordgo.MessageCreate) {
	args := strings.Split(message.Content, " ")
	if len(args) < 2 {
		_, _ = s.ChannelMessageSend(message.ChannelID, "Co chcesz kupić?")
		return
	}
	money := getMoneyForUserId(message.Author.ID)
	switch strings.ToLower(args[1]) {
	case "nick":
		if money.Amount < 500 {
			_, _ = s.ChannelMessageSend(message.ChannelID, "Nie masz wystarczającej ilości pieniędzy :worried:")
			return
		}
		if len(args) < 3 {
			_, _ = s.ChannelMessageSend(message.ChannelID, "Musisz jeszcze powiedzieć jak chcesz się nazywać :worried:")
			return
		}
		newNickname := strings.Join(args[2:], " ")
		newNicknameCleared := clearUsername(newNickname)
		if newNickname != newNicknameCleared {
			_, _ = s.ChannelMessageSend(message.ChannelID, "Nieprawidłowy nick. Możesz spróbować z `"+newNicknameCleared+"`")
			return
		}
		err := s.GuildMemberNickname(message.GuildID, message.Author.ID, newNicknameCleared)
		if err != nil {
			_, _ = s.ChannelMessageSend(message.ChannelID, "Nie mogę zmienić Ci nicku, poinformuj administracje :worried:")
			return
		}
		money.Amount -= 500
		_, err = DbMap.Update(money)
	default:
		_, _ = s.ChannelMessageSend(message.ChannelID, "Nie mam tego w ofercie :worried:")
	}
}
