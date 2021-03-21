package items

import (
	"github.com/bopke/MultisquadDiscordBot/context"
	"github.com/bopke/MultisquadDiscordBot/util"
	"github.com/bopke/MultisquadDiscordBot/vip"
)

func VipHandler(ctx *context.Context, args []string) error {
	embed := util.CreateSimpleEmbed(ctx)
	err := vip.SetVip(ctx.Session, ctx.UserId, 30)
	if err != nil {
		return err
	}
	embed.Description = "Zakup potwierdzony.\nVIP został nadany - miłej zabawy!"
	_, _ = ctx.Session.ChannelMessageSendEmbed(ctx.ChannelId, embed)
	return nil
}
