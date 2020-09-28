package main

import (
	"database/sql"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
	"time"
)

func handleStatusCommand(s *discordgo.Session, message *discordgo.MessageCreate) {
	var linkedUser LinkedUsers
	checkedUser := message.Author
	if len(message.Mentions) > 0 {
		checkedUser = message.Mentions[0]
	}
	log.Println("Sprawdzam status " + checkedUser.Username + "#" + checkedUser.Discriminator + " (" + checkedUser.ID + ") w bazie")
	err := DbMap.SelectOne(&linkedUser, "SELECT expiration_date FROM LinkedUsers WHERE discord_id=?", checkedUser.ID)
	if err == sql.ErrNoRows {
		log.Println("Stwierdzam nieobecność " + checkedUser.Username + "#" + checkedUser.Discriminator + " (" + checkedUser.ID + ")" + " w bazie")
		if len(message.Mentions) > 0 {
			_, _ = s.ChannelMessageSend(message.ChannelID, Locale.StatusNoVipSomeone)
			return
		}
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.StatusNoVip)
		return
	}
	if linkedUser.ExpirationDate.Before(time.Now()) {
		log.Println("Stwierdzam wygaśnięcie statusu vip " + checkedUser.Username + "#" + checkedUser.Discriminator + " (" + checkedUser.ID + ")")
		if len(message.Mentions) > 0 {
			_, _ = s.ChannelMessageSend(message.ChannelID, Locale.StatusExpiredSomeone)
			return
		}
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.StatusExpired)
		return
	}
	statusValid := fmt.Sprintf("%2d.%02d.%4d %2d:%02d", linkedUser.ExpirationDate.Day(), linkedUser.ExpirationDate.Month(), linkedUser.ExpirationDate.Year(), linkedUser.ExpirationDate.Hour(), linkedUser.ExpirationDate.Minute())
	log.Println("Stwierdzam aktywność statusu vip " + checkedUser.Username + "#" + checkedUser.Discriminator + " (" + checkedUser.ID + ") do " + statusValid)
	statusValid = Locale.StatusValid
	if len(message.Mentions) > 0 {
		statusValid = Locale.StatusValidSomeone
	}
	statusValid = strings.Replace(statusValid, "{DAY}", fmt.Sprintf("%d", linkedUser.ExpirationDate.Day()), -1)
	statusValid = strings.Replace(statusValid, "{MONTH}", fmt.Sprintf("%02d", linkedUser.ExpirationDate.Month()), -1)
	statusValid = strings.Replace(statusValid, "{YEAR}", fmt.Sprintf("%d", linkedUser.ExpirationDate.Year()), -1)
	statusValid = strings.Replace(statusValid, "{HOUR}", fmt.Sprintf("%d", linkedUser.ExpirationDate.Hour()), -1)
	statusValid = strings.Replace(statusValid, "{MINUTES}", fmt.Sprintf("%02d", linkedUser.ExpirationDate.Minute()), -1)
	_, _ = s.ChannelMessageSend(message.ChannelID, statusValid)
}
