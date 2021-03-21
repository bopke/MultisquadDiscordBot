package embeds

import (
	"github.com/bopke/MultisquadDiscordBot/context"
	"github.com/bopke/MultisquadDiscordBot/util"
	"github.com/bwmarrin/discordgo"
	"strconv"
)

func BiedaEmbed(ctx *context.Context, amount int) *discordgo.MessageEmbed {
	embed := util.CreateSimpleEmbed(ctx)
	embed.Description = "Sorry mordo, jesteś za biedny.\nPosiadasz tylko " + strconv.Itoa(amount) + "<a:moneta:613020692346175628>"
	return embed
}

func OutOfStockEmbed(ctx *context.Context) *discordgo.MessageEmbed {
	embed := util.CreateSimpleEmbed(ctx)
	embed.Description = "Podany przedmiot się **wyprzedał** :frowning:\nKolejna dostawa z okazji następnego sezonu!"
	return embed
}

func TooEarlyBuyEmbed(ctx *context.Context) *discordgo.MessageEmbed {
	embed := util.CreateSimpleEmbed(ctx)
	embed.Description = "Ten przedmiot możesz kupić tylko raz na 30 dni :("
	return embed
}
