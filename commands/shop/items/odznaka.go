package items

import (
	"github.com/bopke/MultisquadDiscordBot/commands/shop/errors"
	"github.com/bopke/MultisquadDiscordBot/config"
	"github.com/bopke/MultisquadDiscordBot/context"
	"github.com/bopke/MultisquadDiscordBot/util"
)

const odznakaRoleId = "824257782630318181" //"613367448787484703"

func OdznakaHandler(ctx *context.Context, args []string) error {
	if util.HasRoleId(ctx, odznakaRoleId) {
		embed := util.CreateSimpleEmbed(ctx)
		embed.Description = "Już masz odznake :sunglasses:"
		_, _ = ctx.Session.ChannelMessageSendEmbed(ctx.ChannelId, embed)
		return errors.SilentNoSellError
	}
	err := ctx.Session.GuildMemberRoleAdd(config.GuildId, ctx.UserId, odznakaRoleId)
	if err != nil {
		embed := util.CreateSimpleEmbed(ctx)
		embed.Description = "Nie mogłem ustawić Ci rangi :worried:"
		_, _ = ctx.Session.ChannelMessageSendEmbed(ctx.ChannelId, embed)
		return errors.SilentNoSellError
	}
	embed := util.CreateSimpleEmbed(ctx)
	embed.Description = "Zakup potwierdzony.\nLimitowana odznaka sezonowa jest twoja! Gratulacje!"
	_, _ = ctx.Session.ChannelMessageSendEmbed(ctx.ChannelId, embed)
	return nil
}
