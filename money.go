package main

import (
	"github.com/bopke/MultisquadDiscordBot/config"
	"github.com/bopke/MultisquadDiscordBot/database"
	"github.com/bopke/MultisquadDiscordBot/money"
	"github.com/bwmarrin/discordgo"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

var messageMoneyCountMutex = sync.Mutex{}

var moneyClearRequestTime = time.Now()
var moneyClearRequesterId = ""

func handleMoneyZerujCommand(s *discordgo.Session, message *discordgo.MessageCreate, args []string) {
	if !hasPermission(message.Member, config.GuildId, discordgo.PermissionAdministrator) {
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.NoAdminPermission)
		return
	}
	if len(args) == 2 {
		_, _ = s.ChannelMessageSend(message.ChannelID, "Za chwilę wyzerujesz stan monet wszystkich użytkowników. Ich stan zostanie **BEZPOWROTNIE UTRACONY**. Na pewno chcesz to zrobić? W ciągu 20 sekund wpisz `!money zeruj potwierdz`.")
		moneyClearRequesterId = message.Author.ID
		moneyClearRequestTime = time.Now()
	}
	if len(args) == 3 && args[2] == "potwierdz" {
		if message.Author.ID == moneyClearRequesterId && moneyClearRequestTime.Add(time.Second*time.Duration(20)).After(time.Now()) {
			_, _ = s.ChannelMessageSend(message.ChannelID, "Zeruję stan monet. To może chwile zająć. Dam znać gdy skończę.")
			_, err := database.DbMap.Exec("UPDATE Money SET amount=0")
			if err != nil {

			}
			_, _ = s.ChannelMessageSend(message.ChannelID, "Ukończyłem zerowanie.")
		}
		_, _ = s.ChannelMessageSend(message.ChannelID, "Nie ma żadnych aktywnych próśb o wyzerowanie.")
	}
}

func handleMoneyManipulateCommand(s *discordgo.Session, message *discordgo.MessageCreate, args []string) {
	if !hasPermission(message.Member, config.GuildId, discordgo.PermissionAdministrator) && !hasRole(message.Member, "Moderator", config.GuildId) {
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.NoAdminPermission)
		return
	}
	if len(args) != 4 && len(message.Mentions) != 1 {
		_, _ = s.ChannelMessageSend(message.ChannelID, "Poprawne użycie: !monety "+args[1]+" <mention> <ilosc>")
		return
	}

	amount, err := strconv.Atoi(args[3])
	if err != nil {
		_, _ = s.ChannelMessageSend(message.ChannelID, "Poprawne użycie: !monety "+args[1]+" <mention> <ilosc>")
		return
	}
	userMoney := money.GetMoneyForUserId(message.Mentions[0].ID)
	if args[1] == "dodaj" {
		userMoney.Amount += amount
		go logMoneyAdd(s, message.Mentions[0].ID, "added by <@"+message.Author.ID+">", amount, userMoney.Amount)
	} else {
		userMoney.Amount -= amount
		if userMoney.Amount < 0 {
			userMoney.Amount = 0
		}
		go logMoneyAdd(s, message.Mentions[0].ID, "removed by <@"+message.Author.ID+">", -amount, userMoney.Amount)
	}

	messageMoneyCountMutex.Lock()
	defer messageMoneyCountMutex.Unlock()
	_, err = database.DbMap.Update(userMoney)
	if err != nil {
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.UnexpectedApiError)
		return
	}
	_, _ = s.ChannelMessageSend(message.ChannelID, "Nowa ilość monet użytkownika: "+strconv.Itoa(userMoney.Amount))
}

func handleMoneyPrzekazCommand(s *discordgo.Session, message *discordgo.MessageCreate, args []string) {
	if len(args) != 4 || len(message.Mentions) != 1 {
		_, _ = s.ChannelMessageSend(message.ChannelID, "Prawidłowe użycie: !monety wyslij <mention> <ilosc>")
		return
	}
	amount, err := strconv.Atoi(args[3])
	if err != nil {
		_, _ = s.ChannelMessageSend(message.ChannelID, "Prawidłowe użycie: !monety wyslij <mention> <ilosc>")
		return
	}
	if amount < 0 {
		return
	}
	if message.Author.ID == message.Mentions[0].ID {
		//                _,_ = s.ChannelMessageSend(message.ChannelID, "Sam sobie")
		return
	}

	sourceUserMoney := money.GetMoneyForUserId(message.Author.ID)
	if sourceUserMoney.Amount < amount {
		_, _ = s.ChannelMessageSend(message.ChannelID, "Nie masz tyle :worried:")
		return
	}
	destinationUserMoney := money.GetMoneyForUserId(message.Mentions[0].ID)

	messageMoneyCountMutex.Lock()
	defer messageMoneyCountMutex.Unlock()
	sourceUserMoney.Amount -= amount
	destinationUserMoney.Amount += amount
	_, err = database.DbMap.Update(sourceUserMoney)
	if err != nil {
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.UnexpectedApiError)
		return
	}
	_, err = database.DbMap.Update(destinationUserMoney)
	if err != nil {
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.UnexpectedApiError)
		return
	}
	embed := &discordgo.MessageEmbed{
		URL:         "",
		Type:        "",
		Title:       "",
		Description: "<:tak:586340195608166410> <@" + message.Mentions[0].ID + "> otrzymał Twoje " + strconv.Itoa(amount) + " <a:moneta:613020692346175628>",
		Timestamp:   time.Now().Format(time.RFC3339),
		Color:       0x00FF00,
		Footer:      nil,
		Image:       nil,
		Thumbnail:   nil,
		Video:       nil,
		Provider:    nil,
		Author: &discordgo.MessageEmbedAuthor{
			URL:          "",
			Name:         message.Author.Username + "#" + message.Author.Discriminator,
			IconURL:      message.Author.AvatarURL("128"),
			ProxyIconURL: "",
		},
		Fields: nil,
	}
	_, err = s.ChannelMessageSendEmbed(message.ChannelID, embed)
	if err != nil {
		log.Println(err)
	}
}

func handleMoneyCommand(s *discordgo.Session, message *discordgo.MessageCreate) {
	var userMoney *database.Money
	checkedUser := message.Author
	args := strings.Split(message.Content, " ")
	if args[0] != "!mon" && args[0] != "!monety" {
		return
	}
	if len(args) > 1 {
		switch args[1] {
		case "zeruj":
			handleMoneyZerujCommand(s, message, args)
			return
		case "dodaj":
			fallthrough
		case "usun":
			handleMoneyManipulateCommand(s, message, args)
			return
		case "wyslij":
			handleMoneyPrzekazCommand(s, message, args)
			return
		case "top":
			handleBaltopCommand(s, message)
			return
		}
	}
	if len(message.Mentions) > 0 {
		checkedUser = message.Mentions[0]
	}
	log.Println("Sprawdzam pieniadze " + checkedUser.Username + "#" + checkedUser.Discriminator + " (" + checkedUser.ID + ") w bazie")
	userMoney = money.GetMoneyForUserId(checkedUser.ID)
	position, err := money.GetUserMoneyRankPosition(checkedUser.ID)
	if err != nil {
		log.Println(err)
	}
	embed := &discordgo.MessageEmbed{
		URL:         "",
		Type:        "",
		Title:       "",
		Description: "Pozycja w rankingu: " + strconv.Itoa(int(position)),
		Timestamp:   time.Now().Format(time.RFC3339),
		Color:       0x3333FF,
		Footer:      nil,
		Image:       nil,
		Thumbnail:   nil,
		Video:       nil,
		Provider:    nil,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    checkedUser.Username + "#" + checkedUser.Discriminator,
			IconURL: checkedUser.AvatarURL("128"),
		},
		Fields: []*discordgo.MessageEmbedField{{
			Name:  "Stan konta",
			Value: strconv.Itoa(userMoney.Amount) + " <a:moneta:613020692346175628>",
		}},
	}

	_, _ = s.ChannelMessageSendEmbed(message.ChannelID, embed)
}

func handleBaltopCommand(s *discordgo.Session, message *discordgo.MessageCreate) {
	var userMoneys []database.Money
	_, err := database.DbMap.Select(&userMoneys, "SELECT * FROM Money ORDER BY amount DESC LIMIT 10")
	if err != nil {
		log.Println("handleBaltopCommand unable to select from DB")
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.UnexpectedApiError)
		return
	}
	top := ""
	for i, money := range userMoneys {
		top += strconv.Itoa(i+1) + ". <@" + money.UserId + "> - " + strconv.Itoa(money.Amount) + " <a:moneta:613020692346175628>\n"
	}
	position, err := money.GetUserMoneyRankPosition(message.Author.ID)
	if err != nil {
		log.Println(err)
	}
	top += "\n\nTwoja pozycja w rankingu:" + strconv.Itoa(int(position))
	embed := &discordgo.MessageEmbed{
		Description: "",
		Timestamp:   time.Now().Format(time.RFC3339),
		Color:       0xFFFF00,
		Footer:      nil,
		Image:       nil,
		Thumbnail:   nil,
		Video:       nil,
		Provider:    nil,
		Author:      nil,
		Fields: []*discordgo.MessageEmbedField{{
			Name:   "Top 10 serwera",
			Value:  top,
			Inline: false,
		}},
	}
	_, _ = s.ChannelMessageSendEmbed(message.ChannelID, embed)
}

func logMoneyAdd(s *discordgo.Session, userId, reason string, added, whole int) {
	embed := &discordgo.MessageEmbed{
		Title:       "Zdobyte monety",
		Description: "Użytkownik: <@" + userId + ">\nIlość: " + strconv.Itoa(added) + "\nPowód: " + reason + "\nW sumie: " + strconv.Itoa(whole),
		Timestamp:   time.Now().Format(time.RFC3339),
		Color:       0xffff00,
		Fields:      nil,
	}
	_, err := s.ChannelMessageSendEmbed(config.MoneyLogChannelId, embed)
	if err != nil {
		log.Println("logMoneyAdd unable to send embed! ", err)
	}
}

func rankMoneyAdd(roleId string, amount int, after string) {
	members, err := session.GuildMembers(config.GuildId, after, 1000)
	if err != nil {
		log.Println("Error adding money " + err.Error())
		return
	}
	for _, member := range members {
		if hasRoleId(member, roleId, config.GuildId) {
			userMoney := money.GetMoneyForUserId(member.User.ID)
			userMoney.Amount += amount
			_, _ = database.DbMap.Update(userMoney)
			go logMoneyAdd(session, member.User.ID, "has <@&"+roleId+"> role", amount, userMoney.Amount)
		}
	}
	if len(members) == 1000 {
		rankMoneyAdd(roleId, amount, members[999].User.ID)
		return
	}
	embed := &discordgo.MessageEmbed{
		URL:         "",
		Type:        "",
		Title:       "",
		Description: "Użytkownicy z rangą <@&" + roleId + "> dostali " + strconv.Itoa(amount) + " <a:moneta:613020692346175628>",
		Timestamp:   time.Now().Format(time.RFC3339),
		Color:       0x00FF00,
		Footer:      nil,
		Image:       nil,
		Thumbnail:   nil,
		Video:       nil,
		Provider:    nil,
		Author:      nil,
		Fields:      nil,
	}
	_, _ = session.ChannelMessageSendEmbed(config.AnnouncementChannelId, embed)
}
