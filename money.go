package main

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

var messageMoneyCountMutex = sync.Mutex{}
var lastCountedTime = make(map[string]time.Time)

var moneyClearRequestTime = time.Now()
var moneyClearRequesterId = ""

func createMoneyForUserId(userId string) *Money {
	var userMoney Money
	userMoney.Amount = 0
	userMoney.UserId = userId
	err := DbMap.Insert(&userMoney)
	if err != nil {
		return nil
	}
	return &userMoney
}

func getMoneyForUserId(userId string) *Money {
	var userMoney Money
	err := DbMap.SelectOne(&userMoney, "SELECT * FROM Money WHERE user_id=?", userId)
	if err != nil {
		money := createMoneyForUserId(userId)
		if money == nil {
			return nil
		}
		return money
	}
	return &userMoney
}

func messageMoneyCountCS(message *discordgo.MessageCreate) (int, int) {
	messageMoneyCountMutex.Lock()
	defer messageMoneyCountMutex.Unlock()
	lastTime, ok := lastCountedTime[message.Author.ID]
	now := time.Now()
	if !ok {
		lastCountedTime[message.Author.ID] = now
		lastTime = now
	}

	if lastTime != now && lastTime.Add(time.Second*time.Duration(Config.MessageMoneyInterval)).After(now) {
		return -1, -1
	}
	lastCountedTime[message.Author.ID] = now
	userMoney := getMoneyForUserId(message.Author.ID)
	if userMoney == nil {
		return -1, -1
	}
	addedAmount := Config.MessageMoneyMin + rand.Intn(Config.MessageMoneyMax-Config.MessageMoneyMin)
	userMoney.Amount += addedAmount
	_, err := DbMap.Update(userMoney)
	if err != nil {
		log.Println("messageMoneyCountCS cannot update in database! ", err)
		return -1, -1
	}
	return addedAmount, userMoney.Amount
}

func handleMessageMoneyCount(s *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.Bot {
		return
	}
	added, whole := messageMoneyCountCS(message)
	if added == -1 {
		return
	}
	logMoneyAdd(s, message.Author.ID, "Text channel activity", added, whole)
}

func handleMoneyZerujCommand(s *discordgo.Session, message *discordgo.MessageCreate, args []string) {
	if !hasPermission(message.Member, Config.ServerId, discordgo.PermissionAdministrator) {
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
			_, _ = DbMap.Exec("UPDATE Money SET amount=0")
			_, _ = s.ChannelMessageSend(message.ChannelID, "Ukończyłem zerowanie.")
		}
		_, _ = s.ChannelMessageSend(message.ChannelID, "Nie ma żadnych aktywnych próśb o wyzerowanie.")
	}
}

func handleMoneyManipulateCommand(s *discordgo.Session, message *discordgo.MessageCreate, args []string) {
	if !hasPermission(message.Member, Config.ServerId, discordgo.PermissionAdministrator) {
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.NoAdminPermission)
		return
	}
	if len(args) != 4 && len(message.Mentions) != 1 {
		_, _ = s.ChannelMessageSend(message.ChannelID, "Poprawne użycie: !money "+args[1]+" <mention> <ilosc>")
		return
	}

	amount, err := strconv.Atoi(args[3])
	if err != nil {
		_, _ = s.ChannelMessageSend(message.ChannelID, "Poprawne użycie: !money "+args[1]+" <mention> <ilosc>")
		return
	}

	userMoney := getMoneyForUserId(message.Mentions[0].ID)
	if args[1] == "dodaj" {
		userMoney.Amount += amount
	} else {
		userMoney.Amount -= amount
		if userMoney.Amount < 0 {
			userMoney.Amount = 0
		}
	}

	messageMoneyCountMutex.Lock()
	defer messageMoneyCountMutex.Unlock()
	_, err = DbMap.Update(userMoney)
	if err != nil {
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.UnexpectedApiError)
		return
	}
	_, _ = s.ChannelMessageSend(message.ChannelID, "Nowa ilość monet użytkownika: "+strconv.Itoa(userMoney.Amount))
}

func handleMoneyPrzekazCommand(s *discordgo.Session, message *discordgo.MessageCreate, args []string) {
	if len(args) != 4 || len(message.Mentions) != 1 {
		_, _ = s.ChannelMessageSend(message.ChannelID, "Prawidłowe użycie: !money przekaz <mention> <ilosc>")
		return
	}
	amount, err := strconv.Atoi(args[3])
	if err != nil {
		_, _ = s.ChannelMessageSend(message.ChannelID, "Prawidłowe użycie: !money przekaz <mention> <ilosc>")
		return
	}
	sourceUserMoney := getMoneyForUserId(message.Author.ID)
	if sourceUserMoney.Amount < amount {
		_, _ = s.ChannelMessageSend(message.ChannelID, "Nie masz tyle :worried:")
		return
	}
	destinationUserMoney := getMoneyForUserId(message.Mentions[0].ID)

	messageMoneyCountMutex.Lock()
	defer messageMoneyCountMutex.Unlock()
	sourceUserMoney.Amount -= amount
	destinationUserMoney.Amount += amount
	_, err = DbMap.Update(sourceUserMoney)
	if err != nil {
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.UnexpectedApiError)
		return
	}
	_, err = DbMap.Update(destinationUserMoney)
	if err != nil {
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.UnexpectedApiError)
		return
	}
	_, _ = s.ChannelMessageSend(message.ChannelID, "Przekazane.")
}

func handleMoneyCommand(s *discordgo.Session, message *discordgo.MessageCreate) {
	var userMoney *Money
	checkedUser := message.Author
	args := strings.Split(message.Content, " ")
	if len(args) > 1 {
		switch args[1] {
		case "zeruj":
			handleMoneyZerujCommand(s, message, args)
			return
		case "dodaj":
			fallthrough
		case "zabierz":
			handleMoneyManipulateCommand(s, message, args)
			return
		case "przekaz":
			handleMoneyPrzekazCommand(s, message, args)
			return
		}
	}
	if len(message.Mentions) > 0 {
		checkedUser = message.Mentions[0]
	}
	log.Println("Sprawdzam pieniadze " + checkedUser.Username + "#" + checkedUser.Discriminator + " (" + checkedUser.ID + ") w bazie")
	userMoney = getMoneyForUserId(checkedUser.ID)

	result := "Ten użytkownik ma " + strconv.Itoa(userMoney.Amount) + " monet."
	_, _ = s.ChannelMessageSend(message.ChannelID, result)
}

func handleBaltopCommand(s *discordgo.Session, message *discordgo.MessageCreate) {
	var userMoneys []Money
	_, err := DbMap.Select(&userMoneys, "SELECT * FROM Money ORDER BY amount DESC LIMIT 10")
	if err != nil {
		log.Println("handleBaltopCommand unable to select from DB")
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.UnexpectedApiError)
		return
	}
	top := ""
	for i, money := range userMoneys {
		top += strconv.Itoa(i+1) + ". <@" + money.UserId + "> - " + strconv.Itoa(money.Amount) + " monet\n"
	}
	embed := &discordgo.MessageEmbed{
		Timestamp: time.Now().Format(time.RFC3339),
		Color:     0xFFFF00,
		Footer:    nil,
		Image:     nil,
		Thumbnail: nil,
		Video:     nil,
		Provider:  nil,
		Author:    nil,
		Fields: []*discordgo.MessageEmbedField{{
			Name:   "Top 10 bogoli",
			Value:  top,
			Inline: false,
		}},
	}
	_, _ = s.ChannelMessageSendEmbed(message.ChannelID, embed)
}

func logMoneyAdd(s *discordgo.Session, userId, reason string, added, whole int) {
	embed := &discordgo.MessageEmbed{
		Title:       "Zdobyte monety",
		Description: "<@" + userId + ">",
		Timestamp:   time.Now().Format(time.RFC3339),
		Color:       0xffff00,
		Fields: []*discordgo.MessageEmbedField{{
			Name:   reason,
			Value:  "+" + strconv.Itoa(added) + "\nW sumie: " + strconv.Itoa(whole),
			Inline: false,
		}},
	}
	_, err := s.ChannelMessageSendEmbed(Config.MoneyLogChannelId, embed)
	if err != nil {
		log.Println("logMoneyAdd unable to send embed! ", err)
	}
}
