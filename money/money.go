package money

import (
	"github.com/bopke/MultisquadDiscordBot/config"
	"github.com/bopke/MultisquadDiscordBot/context"
	"github.com/bopke/MultisquadDiscordBot/database"
	"github.com/bopke/MultisquadDiscordBot/util"
	"github.com/bwmarrin/discordgo"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

var messageMoneyCountMutex = sync.Mutex{}
var lastCountedTime = make(map[string]time.Time)

const (
	messageMoneyInterval = 90
	messageMoneyMin      = 5
	messageMoneyMax      = 10
	moneyLogChannelId    = "597216492123324442"
)

func createMoneyForUserId(userId string) *database.Money {
	var userMoney database.Money
	userMoney.Amount = 0
	userMoney.UserId = userId
	err := database.DbMap.Insert(&userMoney)
	if err != nil {
		return nil
	}
	return &userMoney
}

func GetMoneyForUserId(userId string) *database.Money {
	var userMoney database.Money
	err := database.DbMap.SelectOne(&userMoney, "SELECT * FROM Money WHERE user_id=?", userId)
	if err != nil {
		money := createMoneyForUserId(userId)
		if money == nil {
			return nil
		}
		return money
	}
	return &userMoney
}

func GetUserMoneyRankPosition(userId string) (int64, error) {
	return database.DbMap.SelectInt("SELECT FIND_IN_SET(amount, (SELECT GROUP_CONCAT( amount ORDER BY amount DESC) FROM Money )) AS rank FROM Money WHERE user_id = ?", userId)
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

	if lastTime != now && lastTime.Add(time.Second*time.Duration(messageMoneyInterval)).After(now) {
		return -1, -1
	}
	lastCountedTime[message.Author.ID] = now
	userMoney := GetMoneyForUserId(message.Author.ID)
	if userMoney == nil {
		return -1, -1
	}
	addedAmount := messageMoneyMin + rand.Intn(messageMoneyMax-messageMoneyMin)
	userMoney.Amount += addedAmount
	_, err := database.DbMap.Update(userMoney)
	if err != nil {
		log.Println("messageMoneyCountCS cannot update in database! ", err)
		return -1, -1
	}
	return addedAmount, userMoney.Amount
}

func HandleMessageMoneyCount(s *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.Bot {
		return
	}
	if message.ChannelID == "580117263815016449" || message.ChannelID == "581950409094987838" {
		return
	}
	added, whole := messageMoneyCountCS(message)
	if added == -1 {
		return
	}
	logMoneyAdd(s, message.Author.ID, "Text channel activity", added, whole)
}

func logMoneyAdd(s *discordgo.Session, userId, reason string, added, whole int) {
	embed := &discordgo.MessageEmbed{
		Title:       "Zdobyte monety",
		Description: "Użytkownik: <@" + userId + ">\nIlość: " + strconv.Itoa(added) + "\nPowód: " + reason + "\nW sumie: " + strconv.Itoa(whole),
		Timestamp:   time.Now().Format(time.RFC3339),
		Color:       0xffff00,
		Fields:      nil,
	}
	_, err := s.ChannelMessageSendEmbed(moneyLogChannelId, embed)
	if err != nil {
		log.Println("logMoneyAdd unable to send embed! ", err)
	}
}

func RankMoneyAdd(session *discordgo.Session, roleId string, amount int, after string) {
	members, err := session.GuildMembers(config.GuildId, after, 1000)
	if err != nil {
		log.Println("Error adding money " + err.Error())
		return
	}
	for _, member := range members {
		if util.HasRoleId(&context.Context{
			Session: session,
			Member:  member,
			GuildId: config.GuildId,
			UserId:  member.User.ID,
		}, roleId) {
			userMoney := GetMoneyForUserId(member.User.ID)
			userMoney.Amount += amount
			_, _ = database.DbMap.Update(userMoney)
			go logMoneyAdd(session, member.User.ID, "has <@&"+roleId+"> role", amount, userMoney.Amount)
		}
	}
	if len(members) == 1000 {
		RankMoneyAdd(session, roleId, amount, members[999].User.ID)
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
