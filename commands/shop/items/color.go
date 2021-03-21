package items

import (
	"github.com/bopke/MultisquadDiscordBot/colors"
	"github.com/bopke/MultisquadDiscordBot/context"
	"github.com/bopke/MultisquadDiscordBot/util"
	"strings"
)

func ColorHandler(ctx *context.Context, args []string) error {
	if len(args) < 1 {
		embed := util.CreateSimpleEmbed(ctx)
		embed.Title = "Nieprawidłowe użycie"
		embed.Description = "Poprawne użycie !kup kolor <nazwa koloru>"
		_, _ = ctx.Session.ChannelMessageSendEmbed(ctx.ChannelId, embed)
		return nil
	}
	embed := util.CreateSimpleEmbed(ctx)
	err := colors.SetUserColor(ctx.Session, ctx.UserId, strings.Join(args, " "), 30)
	if err != nil {
		if err == colors.NoSuchColorError {
			embed := util.CreateSimpleEmbed(ctx)
			embed.Title = "Nieprawidłowe użycie"
			embed.Description = "Nie ma takiego koloru :worried:"
			return nil
		}
		return err
	}
	embed.Description = "Zakup potwierdzony."
	_, _ = ctx.Session.ChannelMessageSendEmbed(ctx.ChannelId, embed)
	return nil
}
