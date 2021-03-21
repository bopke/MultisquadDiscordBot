package items

import (
	"github.com/bopke/MultisquadDiscordBot/commands/shop/errors"
	"github.com/bopke/MultisquadDiscordBot/config"
	"github.com/bopke/MultisquadDiscordBot/context"
	"github.com/bopke/MultisquadDiscordBot/util"
)

const flexerRoleId = "823190977077706772" //"613367448787484703"

func FlexerHandler(ctx *context.Context, args []string) error {
	if util.HasRoleId(ctx, flexerRoleId) {
		embed := util.CreateSimpleEmbed(ctx)
		embed.Description = "Już jesteś flexerem :sunglasses:"
		_, _ = ctx.Session.ChannelMessageSendEmbed(ctx.ChannelId, embed)
		return errors.SilentNoSellError
	}
	err := ctx.Session.GuildMemberRoleAdd(config.GuildId, ctx.UserId, flexerRoleId)
	if err != nil {
		embed := util.CreateSimpleEmbed(ctx)
		embed.Description = "Nie mogłem ustawić Ci rangi :worried:"
		_, _ = ctx.Session.ChannelMessageSendEmbed(ctx.ChannelId, embed)
		return errors.SilentNoSellError
	}
	embed := util.CreateSimpleEmbed(ctx)
	embed.Description = "Zakup potwierdzony.\nGratulacje, wydałeś właśnie 100k na rangę, która nic nie daje!\n**PRAWDZIWY F L E X**"
	_, _ = ctx.Session.ChannelMessageSendEmbed(ctx.ChannelId, embed)
	return nil
}
