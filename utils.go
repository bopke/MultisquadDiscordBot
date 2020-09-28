package main

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
)

func addReactionToImage(message *discordgo.Message) {
	for _, reactionChannel := range Config.ReactionChannelsId {
		if reactionChannel == message.ChannelID {
			if len(message.Attachments) != 0 || len(message.Embeds) != 0 || strings.Contains(message.Content, "http://") || strings.Contains(message.Content, "https://") {
				err := session.MessageReactionAdd(message.ChannelID, message.ID, Config.ReactionId)
				if err != nil {
					log.Println("Błąd dodawania reakcji do wiadomosci: ", err)
				}
			}
			break
		}
	}
}
