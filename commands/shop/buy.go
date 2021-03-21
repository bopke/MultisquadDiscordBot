package shop

import (
	"github.com/bopke/MultisquadDiscordBot/commands/errors"
	"github.com/bopke/MultisquadDiscordBot/commands/shop/embeds"
	errors2 "github.com/bopke/MultisquadDiscordBot/commands/shop/errors"
	shopItems "github.com/bopke/MultisquadDiscordBot/commands/shop/items"
	"github.com/bopke/MultisquadDiscordBot/context"
	"github.com/bopke/MultisquadDiscordBot/database"
	"github.com/bopke/MultisquadDiscordBot/money"
	"log"
	"strings"
	"sync"
	"time"
)

type itemInfo struct {
	Price        int
	ItemHandler  func(ctx *context.Context, args []string) error
	Limit        int
	LimitPerUser bool
}

var items = map[string]itemInfo{
	"flexer": {
		Price:       100000,
		ItemHandler: shopItems.FlexerHandler,
		Limit:       -1,
	},
	"nick": {
		Price:       500,
		ItemHandler: shopItems.NickHandler,
		Limit:       -1,
	},
	"nitro": {
		Price:        90000,
		ItemHandler:  shopItems.NitroHandler,
		Limit:        3,
		LimitPerUser: true,
	},
	"vip": {
		Price:        65000,
		ItemHandler:  shopItems.VipHandler,
		Limit:        5,
		LimitPerUser: true,
	},
	"kolor": {
		Price:       15000,
		ItemHandler: shopItems.ColorHandler,
		Limit:       -1,
	},
}

var shopMutex = sync.Mutex{}

func BuyCommand(ctx *context.Context, args []string) (string, error) {
	if len(args) < 1 {
		return "Co chcesz kupić?", errors.IncorrectUsageError
	}
	balance := money.GetMoneyForUserId(ctx.UserId)
	item, exists := items[strings.ToLower(args[0])]
	if !exists {
		return "Nie mam tego w ofercie :worried:", errors.IncorrectUsageError
	}
	if balance.Amount < item.Price {
		_, _ = ctx.Session.ChannelMessageSendEmbed(ctx.ChannelId, embeds.BiedaEmbed(ctx, balance.Amount))
		return "", nil
	}
	shopMutex.Lock()
	defer shopMutex.Unlock()

	if item.Limit > 0 {
		var shopLogs []database.ShopLog
		_, err := database.DbMap.Select(&shopLogs, "SELECT * FROM ShopLogs WHERE item = \"nitro\"")
		if err != nil {
			log.Panicln(err)
		}
		if len(shopLogs) >= item.Limit {
			_, _ = ctx.Session.ChannelMessageSendEmbed(ctx.ChannelId, embeds.OutOfStockEmbed(ctx))
			return "", nil
		}
		if item.LimitPerUser {
			var shopLog database.ShopLog
			err = database.DbMap.SelectOne(&shopLog, "SELECT * FROM ShopLogs WHERE discord_id = ? AND item=\"nitro\" ORDER BY date DESC", ctx.UserId)
			if err == nil {
				if time.Now().Before(shopLog.Date.Add(30 * 24 * time.Hour)) {
					_, _ = ctx.Session.ChannelMessageSendEmbed(ctx.ChannelId, embeds.TooEarlyBuyEmbed(ctx))
					return "", nil
				}
			}
		}
	}

	balance.Amount -= item.Price
	err := item.ItemHandler(ctx, args[1:])
	if err == nil {
		_, err = database.DbMap.Update(balance)
		if err != nil {
			return "", database.DatabaseError
		}
		shopLog := database.ShopLog{
			DiscordId: ctx.UserId,
			Item:      strings.ToLower(args[0]),
			Price:     item.Price,
			Date:      time.Now(),
		}
		_ = database.DbMap.Insert(&shopLog)

	} else if err == errors2.SilentNoSellError {
		return "", nil
	} else {
		return "", err
	}
	/*	switch () {
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

			if balance.Amount < 65000 {
				_, _ = s.ChannelMessageSendEmbed(message.ChannelID, biedaEmbed(message, balance.Amount))
				return
			}
			balance.Amount -= 65000
			embed := createSimpleEmbed(message)
			message.ID = "0"
			message.Mentions = []*discordgo.User{{ID: message.Author.ID, Username: message.Author.Username, Discriminator: message.Author.Discriminator}}
			message.Content = "!vip <@" + message.Author.ID + "> 30"
			message.Author.ID = "320573515755683840"
			go handleVipCommand(s, message)
			embed.Description = "Zakup potwierdzony.\nVIP został nadany - miłej zabawy!"
			_, _ = s.ChannelMessageSendEmbed(message.ChannelID, embed)
			_, _ = DbMap.Update(balance)
			shopLog = ShopLog{
				DiscordId: message.Author.ID,
				Item:      "vip",
				Price:     65000,
				Date:      time.Now(),
			}
			_ = DbMap.Insert(&shopLog)
		case "kolor":
			if balance.Amount < 15000 {
				_, _ = s.ChannelMessageSendEmbed(message.ChannelID, biedaEmbed(message, balance.Amount))
				return
			}
			var roleId string
			err := s.GuildMemberRoleAdd(message.GuildID, message.Author.ID, roleId)
			if err != nil {
				embed := createSimpleEmbed(message)
				embed.Description = "Nie mogłem nadać Ci roli :("
				_, _ = s.ChannelMessageSendEmbed(message.ChannelID, embed)
				log.Println(err)
				return
			}
			_, _ = database.DbMap.Update(balance)
			shopLog := database.ShopLog{
				DiscordId: ctx.UserId,
				Item:      "color",
				Price:     15000,
				Date:      time.Now(),
			}
			embed := util.CreateSimpleEmbed(ctx)
			embed.Description = "Zakup potwierdzony.\n"
			_, _ = ctx.Session.ChannelMessageSendEmbed(ctx.ChannelId, embed)

			_ = database.DbMap.Insert(&shopLog)
		default:
		}
	*/
	return "", nil
}

func sendShoplog(shopLog database.ShopLog) {

}
