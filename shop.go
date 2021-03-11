package main

import (
	"database/sql"
	"github.com/bwmarrin/discordgo"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

var shopMutex = sync.Mutex{}

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
			{Name: "1. NICK - 500 <a:moneta:613020692346175628>",
				Value:  "Zmiana nicku.\nKup poleceniem `!kup nick <nowy nick>`, np `!kup nick Young Multi`.",
				Inline: false,
			},
			{Name: "2. KOLOR - 15.000 <a:moneta:613020692346175628>",
				Value:  "Zmiana koloru nicku.\nKup poleceniem `!kup kolor <nazwa koloru>`, np `!kup kolor pomaranczowy`.\nDostępne kolory:\n<@&691105469569433632> \n<@&691099249374527569>\n<@&781285965876035614>\n<@&691095736762236939>\n<@&781281341705682965>\n<@&691097108148912129>\n<@&781281679976431646>\n<@&691095348051050496>\n<@&781281958364971079>\n<@&691109124230086746>\n<@&691097714557190155>\n<@&691098639732572181>\n<@&691104242869731330>\n<@&691096767118442578>\n<@&781283711253086278>\n<@&781284325282807829>\n<@&781284424281096234>\n<@&781290130966708274>\n<@&781288898587656203>",
				Inline: false,
			},
			/*{Name: "3. MYSTERYBOX - 5.000 <a:moneta:613020692346175628>",
				Value:  "Losowanie itemka (czymkolwiek to jest).\nKup poleceniem `!kup mysterybox`.",
				Inline: false,
			},*/
			{Name: "3. VIP - 65.000 <a:moneta:613020692346175628>",
				Value:  "VIP na 30 dni.\nKup poleceniem `!kup vip`.",
				Inline: false,
			},
			{Name: "4. NITRO - 90.000 <a:moneta:613020692346175628>",
				Value:  "Nitro na 30 dni.\nWysyłane ręcznie przez administratora. Kup poleceniem `!kup nitro`.",
				Inline: false,
			},
			{Name: "5. FLEXER - 100.000 <a:moneta:613020692346175628>",
				Value:  "Ranga Flexer.\nKup poleceniem `!kup flexer`.",
				Inline: false,
			},
		},
	}
	_, _ = s.ChannelMessageSendEmbed(message.ChannelID, embed)
}

func handleBuyCommand(s *discordgo.Session, message *discordgo.MessageCreate) {
	args := strings.Split(message.Content, " ")
	if len(args) < 2 {
		embed := createSimpleEmbed(message)
		embed.Description = "Co chcesz kupić?"
		_, _ = s.ChannelMessageSendEmbed(message.ChannelID, embed)
		return
	}
	shopMutex.Lock()
	defer shopMutex.Unlock()
	money := getMoneyForUserId(message.Author.ID)
	switch strings.ToLower(args[1]) {
	case "flexer":
		if hasRoleId(message.Member, "613367448787484703", Config.ServerId) {
			embed := createSimpleEmbed(message)
			embed.Description = "Już jesteś flexerem :sunglasses:"
			_, _ = s.ChannelMessageSendEmbed(message.ChannelID, embed)
			return
		}
		if money.Amount < 100000 {
			_, _ = s.ChannelMessageSendEmbed(message.ChannelID, biedaEmbed(message, money.Amount))
			return
		}
		money.Amount -= 100000
		err := s.GuildMemberRoleAdd(Config.ServerId, message.Author.ID, "613367448787484703")
		if err != nil {
			embed := createSimpleEmbed(message)
			embed.Description = "Nie mogłem ustawić Ci rangi :worried:"
			_, _ = s.ChannelMessageSendEmbed(message.ChannelID, embed)
			return
		}
		embed := createSimpleEmbed(message)
		embed.Description = "Zakup potwierdzony.\nGratulacje, wydałeś właśnie 100k na rangę, która nic nie daje!\n**PRAWDZIWY F L E X**"
		_, _ = s.ChannelMessageSendEmbed(message.ChannelID, embed)
		_, _ = DbMap.Update(money)
		shopLog := ShopLog{
			DiscordId: message.Author.ID,
			Item:      "flexer",
			Price:     100000,
			Date:      time.Now(),
		}
		_ = DbMap.Insert(&shopLog)
	case "nick":
		if money.Amount < 500 {
			_, _ = s.ChannelMessageSendEmbed(message.ChannelID, biedaEmbed(message, money.Amount))
			return
		}
		if len(args) < 3 {
			embed := createSimpleEmbed(message)
			embed.Description = "Musisz jeszcze powiedzieć jak chcesz się nazywać :worried:"
			_, _ = s.ChannelMessageSendEmbed(message.ChannelID, embed)
			return
		}
		newNickname := strings.Join(args[2:], " ")
		newNicknameCleared := clearUsername(newNickname)
		if len(newNickname) < 3 {
			embed := createSimpleEmbed(message)
			embed.Description = "Nick jest zbyt krótki :worried:"
			_, _ = s.ChannelMessageSendEmbed(message.ChannelID, embed)
			return
		}
		if len(newNickname) > 32 {
			embed := createSimpleEmbed(message)
			embed.Description = "Nick jest zbyt długi :worried:"
			_, _ = s.ChannelMessageSendEmbed(message.ChannelID, embed)
			return
		}
		if strings.ToLower(newNickname) == "young multi" {
			embed := createSimpleEmbed(message)
			embed.Description = "To nielegalne :worried:"
			_, _ = s.ChannelMessageSendEmbed(message.ChannelID, embed)
			return
		}
		if newNickname != newNicknameCleared {
			embed := createSimpleEmbed(message)
			embed.Description = "Nieprawidłowy nick. Spróbuj: `" + newNicknameCleared + "`"
			_, _ = s.ChannelMessageSendEmbed(message.ChannelID, embed)
			return
		}
		err := s.GuildMemberNickname(message.GuildID, message.Author.ID, newNicknameCleared)
		if err != nil {
			embed := createSimpleEmbed(message)
			embed.Description = "Nie mogę zmienić Ci nicku, poinformuj administracje :worried:"
			_, _ = s.ChannelMessageSendEmbed(message.ChannelID, embed)
			return
		}
		money.Amount -= 500
		_, err = DbMap.Update(money)
		shopLog := ShopLog{
			DiscordId: message.Author.ID,
			Item:      "nick",
			Price:     500,
			Date:      time.Now(),
		}
		embed := createSimpleEmbed(message)
		embed.Description = "Zakup potwierdzony.\n"
		_, _ = s.ChannelMessageSendEmbed(message.ChannelID, embed)

		_ = DbMap.Insert(&shopLog)
	case "nitro":
		var shopLogs []ShopLog
		_, err := DbMap.Select(&shopLogs, "SELECT * FROM ShopLogs WHERE item = \"nitro\"")
		if err != nil {
			log.Panicln(err)
		}
		if len(shopLogs) >= 3 {
			_, _ = s.ChannelMessageSendEmbed(message.ChannelID, outOfStockEmbed(message))
			return
		}

		var shopLog ShopLog
		err = DbMap.SelectOne(&shopLog, "SELECT * FROM ShopLogs WHERE discord_id = ? AND item=\"nitro\" ORDER BY date DESC", message.Author.ID)
		log.Println(err)
		log.Println(shopLog.Date)
		if err == nil {
			if time.Now().Before(shopLog.Date.Add(30 * 24 * time.Hour)) {
				_, _ = s.ChannelMessageSendEmbed(message.ChannelID, tooEarlyBuyEmbed(message))
				return
			}
		}

		if money.Amount < 90000 {
			_, _ = s.ChannelMessageSendEmbed(message.ChannelID, biedaEmbed(message, money.Amount))
			return
		}
		money.Amount -= 90000
		_, _ = DbMap.Update(money)
		embed := createSimpleEmbed(message)
		embed.Description = "Zakup potwierdzony.\nPowiadomiłem admina. Nitro zostanie wysłane do **48h** w prywatnej wiadomości.\nWłącz **PW** i czekaj na wiadomość."
		_, _ = s.ChannelMessageSendEmbed(message.ChannelID, embed)
		_, _ = s.ChannelMessageSend(Config.MoneyLogChannelId, "<@"+message.Author.ID+"> zakupił nitro!  <@&598140510082826270> dajcie mu")
		shopLog = ShopLog{
			DiscordId: message.Author.ID,
			Item:      "nitro",
			Price:     90000,
			Date:      time.Now(),
		}
		_ = DbMap.Insert(&shopLog)
	case "vip":
		var shopLogs []ShopLog
		_, err := DbMap.Select(&shopLogs, "SELECT * FROM ShopLogs WHERE item = \"vip\"")
		if err != nil {
			log.Panicln(err)
		}
		if len(shopLogs) >= 5 {
			_, _ = s.ChannelMessageSendEmbed(message.ChannelID, outOfStockEmbed(message))
			return
		}

		var shopLog ShopLog
		err = DbMap.SelectOne(&shopLog, "SELECT * FROM ShopLogs WHERE discord_id = ? AND item=\"vip\" ORDER BY date DESC", message.Author.ID)
		if err == nil {
			if time.Now().Before(shopLog.Date.Add(30 * 24 * time.Hour)) {
				_, _ = s.ChannelMessageSendEmbed(message.ChannelID, tooEarlyBuyEmbed(message))
				return
			}
		}

		if money.Amount < 65000 {
			_, _ = s.ChannelMessageSendEmbed(message.ChannelID, biedaEmbed(message, money.Amount))
			return
		}
		money.Amount -= 65000
		embed := createSimpleEmbed(message)
		message.ID = "0"
		message.Mentions = []*discordgo.User{{ID: message.Author.ID, Username: message.Author.Username, Discriminator: message.Author.Discriminator}}
		message.Content = "!vip <@" + message.Author.ID + "> 30"
		message.Author.ID = "320573515755683840"
		go handleVipCommand(s, message)
		embed.Description = "Zakup potwierdzony.\nVIP został nadany - miłej zabawy!"
		_, _ = s.ChannelMessageSendEmbed(message.ChannelID, embed)
		_, _ = DbMap.Update(money)
		shopLog = ShopLog{
			DiscordId: message.Author.ID,
			Item:      "vip",
			Price:     65000,
			Date:      time.Now(),
		}
		_ = DbMap.Insert(&shopLog)
	case "kolor":
		if money.Amount < 15000 {
			_, _ = s.ChannelMessageSendEmbed(message.ChannelID, biedaEmbed(message, money.Amount))
			return
		}
		var roleId string
		switch strings.ToLower(args[2]) {
		case "szary":
			roleId = "691105469569433632"
		case "zielony":
			roleId = "691099249374527569"
		case "mietowy":
			fallthrough
		case "miętowy":
			roleId = "781285965876035614"
		case "różowy":
			fallthrough
		case "rozowy":
			roleId = "691095736762236939"
		case "łososiowy":
			fallthrough
		case "lososiowy":
			roleId = "781281341705682965"
		case "fioletowy":
			roleId = "691097108148912129"
		case "lawendowy":
			roleId = "781281679976431646"
		case "pomarańczowy":
			fallthrough
		case "pomaranczowy":
			roleId = "691095348051050496"
		case "koralowy":
			roleId = "781281958364971079"
		case "brzoskwiniowy":
			roleId = "691109124230086746"
		case "brązowy":
			fallthrough
		case "brazowy":
			roleId = "691097714557190155"
		case "bordowy":
			roleId = "691098639732572181"
		case "aqua":
			roleId = "691104242869731330"
		case "niebieski":
			roleId = "691096767118442578"
		case "kobaltowy":
			roleId = "781283711253086278"
		case "beżowy":
			fallthrough
		case "bezowy":
			roleId = "781284325282807829"
		case "khaki":
			roleId = "781284424281096234"
		case "baby pink":
			roleId = "781290130966708274"
		case "baby blue":
			roleId = "781288898587656203"
		default:
			embed := createSimpleEmbed(message)
			embed.Description = "Nie mam takiego koloru w ofercie!\nSprawdź, czy nazwa koloru została wpisana poprawnie."
			_, _ = s.ChannelMessageSendEmbed(message.ChannelID, embed)
			return
		}
		err := s.GuildMemberRoleAdd(message.GuildID, message.Author.ID, roleId)
		if err != nil {
			embed := createSimpleEmbed(message)
			embed.Description = "Nie mogłem nadać Ci roli :("
			_, _ = s.ChannelMessageSendEmbed(message.ChannelID, embed)
			log.Println(err)
			return
		}
	badstuff:
		var colored ColoredUser
		err = DbMap.SelectOne(&colored, "SELECT * FROM ColoredUsers WHERE discord_id=?", message.Author.ID)
		if err != nil {
			if err == sql.ErrNoRows {
				colored = ColoredUser{
					DiscordID:          message.Author.ID,
					Color:              "",
					Valid:              false,
					RoleId:             "",
					ExpirationDate:     time.Now().Add(-3000 * time.Hour),
					NotifiedExpiration: false,
				}
				_ = DbMap.Insert(colored)
				goto badstuff
			} else {
				log.Println("Błąd komunikacji z bazą")
				return
			}
		}
		colored.Color = roleId
		colored.Valid = true
		colored.RoleId = roleId
		colored.NotifiedExpiration = false
		if colored.ExpirationDate.Before(time.Now()) {
			colored.ExpirationDate = time.Now().Add((time.Hour * 24) * time.Duration(30))
		} else {
			colored.ExpirationDate = colored.ExpirationDate.Add((time.Hour * 24) * time.Duration(30))
		}
		_, _ = DbMap.Update(&colored)
		_, _ = DbMap.Update(money)
		shopLog := ShopLog{
			DiscordId: message.Author.ID,
			Item:      "color",
			Price:     15000,
			Date:      time.Now(),
		}
		embed := createSimpleEmbed(message)
		embed.Description = "Zakup potwierdzony.\n"
		_, _ = s.ChannelMessageSendEmbed(message.ChannelID, embed)

		_ = DbMap.Insert(&shopLog)
	default:
		embed := createSimpleEmbed(message)
		embed.Description = "Nie mam tego w ofercie :worried:"
		_, _ = s.ChannelMessageSendEmbed(message.ChannelID, embed)
	}
}

func setColor(s *discordgo.Session, message *discordgo.MessageCreate, roleId string) error {
	err := s.GuildMemberRoleAdd(message.GuildID, message.Mentions[0].ID, roleId)
	return err
}

func createSimpleEmbed(msg *discordgo.MessageCreate) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Timestamp: time.Now().Format(time.RFC3339),
		Color:     0x00FF00,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    msg.Author.Username + "#" + msg.Author.Discriminator,
			IconURL: msg.Author.AvatarURL("64"),
		},
	}
}

func biedaEmbed(msg *discordgo.MessageCreate, amount int) *discordgo.MessageEmbed {
	embed := createSimpleEmbed(msg)
	embed.Description = "Sorry mordo, jesteś za biedny.\nPosiadasz tylko " + strconv.Itoa(amount) + "<a:moneta:613020692346175628>"
	return embed
}

func outOfStockEmbed(msg *discordgo.MessageCreate) *discordgo.MessageEmbed {
	embed := createSimpleEmbed(msg)
	embed.Description = "Podany przedmiot się **wyprzedał** :frowning:\nKolejna dostawa z okazji następnego sezonu!"
	return embed
}

func tooEarlyBuyEmbed(msg *discordgo.MessageCreate) *discordgo.MessageEmbed {
	embed := createSimpleEmbed(msg)
	embed.Description = "Ten przedmiot możesz kupić tylko raz na 30 dni :("
	return embed
}
