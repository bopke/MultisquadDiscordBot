package main

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
	"sync"
	"time"
)

type ReportDMData struct {
	Stage         int
	reportMessage string
	StageMessages [3]*discordgo.Message
}

var reportDMChannelsStages = map[string]*ReportDMData{}
var reportDMChannelsStagesMutex sync.Mutex

func handleDMReactions(s *discordgo.Session, reaction *discordgo.MessageReactionAdd, channel *discordgo.Channel) {
	reportDMChannelsStagesMutex.Lock()
	defer reportDMChannelsStagesMutex.Unlock()
	reportDMData, ok := reportDMChannelsStages[reaction.UserID]
	if !ok {
		return
	}
	log.Println("Handling DM report reaction")
	switch reportDMData.Stage {
	case 4:
		if reaction.Emoji.ID != "✅" && reaction.Emoji.Name != "✅" {
			_, err := s.ChannelMessageSend(reaction.ChannelID, Locale.ReportDeclinedMessage)
			if err != nil {
				log.Println("handleHMMessage Unable to send DM response on stage 4! ", err)
			}
			delete(reportDMChannelsStages, reaction.UserID)
			return
		}
		_, err := s.ChannelMessageSend(reaction.ChannelID, Locale.ReportConfirmedMessage)
		if err != nil {
			log.Println("handleHMMessage Unable to send DM response on stage 4! ", err)
		}
		sendFullDMReport(reportDMData)
		delete(reportDMChannelsStages, reaction.UserID)
	}
}

func handleDMMessages(s *discordgo.Session, message *discordgo.MessageCreate, channel *discordgo.Channel) {
	reportDMChannelsStagesMutex.Lock()
	defer reportDMChannelsStagesMutex.Unlock()
	reportDMData, ok := reportDMChannelsStages[message.Author.ID]
	if !ok {
		return
	}
	log.Println("Handling DM report message")
	switch reportDMData.Stage {
	case 3:
		reportDMData.StageMessages[2] = message.Message
		reportDMData.Stage = 4
		msg, err := s.ChannelMessageSend(message.ChannelID, Locale.ReportStage4Message)
		if err != nil {
			log.Println("handleHMMessage Unable to send DM response on stage 3! ", err)
		}
		_ = s.MessageReactionAdd(message.ChannelID, msg.ID, "✅")
		_ = s.MessageReactionAdd(message.ChannelID, msg.ID, "❌")
	case 2:
		reportDMData.StageMessages[1] = message.Message
		reportDMData.Stage = 3
		_, err := s.ChannelMessageSend(message.ChannelID, Locale.ReportStage3Message)
		if err != nil {
			log.Println("handleHMMessage Unable to send DM response on stage 2! ", err)
		}
	case 1:
		reportDMData.StageMessages[0] = message.Message
		reportDMData.Stage = 2
		_, err := s.ChannelMessageSend(message.ChannelID, Locale.ReportStage2Message)
		if err != nil {
			log.Println("handleHMMessage Unable to send DM response on stage 1! ", err)
		}
	case 0:
		if !strings.Contains(message.Content, "<@") {
			_, _ = s.ChannelMessageSend(message.ChannelID, "No gosciu miales oznaczyć :worried:")
			return
		}
		reportDMData.reportMessage = message.Content
		reportDMData.Stage = 1
		_, err := s.ChannelMessageSend(message.ChannelID, Locale.ReportStage1Message)
		if err != nil {
			log.Println("handleHMMessage Unable to send DM response on stage 0! ", err)
		}
	}
}

func handleReportCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	err := s.ChannelMessageDelete(m.ChannelID, m.ID)
	if err != nil {
		log.Println("handleReportMessage Unable to delete message! ", err)
	}

	reportDMChannelsStagesMutex.Lock()
	log.Println("Tworze nowy report")
	defer reportDMChannelsStagesMutex.Unlock()
	args := strings.Split(m.Content, " ")
	if len(m.Mentions) == 0 {
		dmChannel, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			msg, err := s.ChannelMessageSend(m.ChannelID, Locale.ErrorCreatingDMChannel)
			if err == nil {
				time.Sleep(20 * time.Second)
				err = s.ChannelMessageDelete(msg.ChannelID, msg.ID)
				if err != nil {
					log.Println("handleReportMessage Unable to delete message! ", err)
				}

			}
			return
		}
		_, _ = s.ChannelMessageSendComplex(dmChannel.ID, &discordgo.MessageSend{
			Content: Locale.ReportStage0Message1,
			Embed: &discordgo.MessageEmbed{
				Image: &discordgo.MessageEmbedImage{
					URL: "https://cdn.discordapp.com/attachments/599975835717468170/715583005674045470/report1.gif",
				},
			},
		})
		_, err = s.ChannelMessageSendComplex(dmChannel.ID, &discordgo.MessageSend{
			Content: Locale.ReportStage0Message2,
			Embed: &discordgo.MessageEmbed{
				Image: &discordgo.MessageEmbedImage{
					URL: "https://cdn.discordapp.com/attachments/599975835717468170/715583017397125180/developer.gif",
				},
			},
		})
		if err != nil {
			msg, err := s.ChannelMessageSend(m.ChannelID, strings.ReplaceAll(Locale.ErrorCreatingDMChannel, "{MENTION}", m.Author.Mention()))
			if err == nil {
				time.Sleep(20 * time.Second)
				err = s.ChannelMessageDelete(msg.ChannelID, msg.ID)
				if err != nil {
					log.Println("handleReportMessage Unable to delete message! ", err)
				}
			}
			return
		}
		reportDMChannelsStages[m.Author.ID] = &ReportDMData{
			Stage:         0,
			StageMessages: [3]*discordgo.Message{},
			reportMessage: "",
		}
		return
	}
	dmChannel, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		msg, err := s.ChannelMessageSend(m.ChannelID, Locale.ErrorCreatingDMChannel)
		if err == nil {
			time.Sleep(20 * time.Second)
			err = s.ChannelMessageDelete(msg.ChannelID, msg.ID)
			if err != nil {
				log.Println("handleReportMessage Unable to delete message! ", err)
			}
		}
		return
	}
	_, err = s.ChannelMessageSend(dmChannel.ID, Locale.ReportStage1Message)
	if err != nil {
		msg, err := s.ChannelMessageSend(m.ChannelID, strings.ReplaceAll(Locale.ErrorCreatingDMChannel, "{MENTION}", m.Author.Mention()))
		if err == nil {
			time.Sleep(20 * time.Second)
			err = s.ChannelMessageDelete(msg.ChannelID, msg.ID)
			if err != nil {
				log.Println("handleReportMessage Unable to delete message! ", err)
			}
		}
		return
	}
	reportDMChannelsStages[m.Author.ID] = &ReportDMData{
		Stage:         1,
		StageMessages: [3]*discordgo.Message{},
		reportMessage: strings.Join(args[1:], " "),
	}
}

func sendFullDMReport(data *ReportDMData) {
	embed := &discordgo.MessageEmbed{
		Title:       "Nowe zgłoszenie: ",
		Description: data.reportMessage,
		Timestamp:   time.Now().Format(time.RFC3339),
		Color:       0xFF0000,
		Author: &discordgo.MessageEmbedAuthor{
			Name: data.StageMessages[0].Author.Username + "#" + data.StageMessages[0].Author.Discriminator,
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: "**Kanał zgłoszenia**",
			},
			{
				Name: "**Powód**",
			},
			{
				Name: "**Dowody**",
			},
		},
	}
	attachments := ""
	for i, stageMessage := range data.StageMessages {
		info := stageMessage.Content + "\n"
		for i, attachment := range stageMessage.Attachments {
			if i == 0 {
				info += "\n\n**Załączniki:**\n"
			}
			attachments += attachment.URL + "\n"
			info += attachment.URL + "\n"
		}
		embed.Fields[i].Value = info
	}
	_, err := session.ChannelMessageSend(Config.ReportsChannelId, "<@&610997643413291008> <@&320577390965424138>")
	if err != nil {
		log.Println("sendFullDMReport unable to send pings! ", err)
	}
	_, err = session.ChannelMessageSendEmbed(Config.ReportsChannelId, embed)
	if err != nil {
		_, _ = session.ChannelMessageSend(Config.ReportsChannelId, "Dostałem zgłoszenie, ale było zbyt długie aby wrzucić je w jednym embedzie. Wrzucam tekstowo w oddzielnych wiadomościach.")
		_, _ = session.ChannelMessageSend(Config.ReportsChannelId, embed.Title+"\n"+embed.Description+"\nAutor:\n"+embed.Author.Name)
		for i, stageMessage := range data.StageMessages {
			_, _ = session.ChannelMessageSend(Config.ReportsChannelId, embed.Fields[i].Name)
			_, _ = session.ChannelMessageSend(Config.ReportsChannelId, stageMessage.Content)
			for i, attachment := range stageMessage.Attachments {
				if i == 0 {
					_, _ = session.ChannelMessageSend(Config.ReportsChannelId, "**Załączniki:**")
				}
				_, _ = session.ChannelMessageSend(Config.ReportsChannelId, attachment.URL)
			}
		}
		//		log.Println(err)
	} else {
		if len(attachments) != 0 {
			_, _ = session.ChannelMessageSend(Config.ReportsChannelId, "Wszystkie załączniki:\n"+attachments)
		}
	}
}
