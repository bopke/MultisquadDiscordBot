package items

import (
	"github.com/bopke/MultisquadDiscordBot/commands/shop/errors"
	"github.com/bopke/MultisquadDiscordBot/context"
	"github.com/bopke/MultisquadDiscordBot/nicks"
	"github.com/bopke/MultisquadDiscordBot/util"
	"strings"
)

func NickHandler(ctx *context.Context, args []string) error {
	if len(args) < 1 {
		embed := util.CreateSimpleEmbed(ctx)
		embed.Description = "Musisz jeszcze powiedzieć jak chcesz się nazywać :worried:"
		_, _ = ctx.Session.ChannelMessageSendEmbed(ctx.ChannelId, embed)
		return errors.SilentNoSellError
	}
	newNickname := strings.Join(args, " ")
	newNicknameCleared := nicks.ClearUsername(newNickname)
	if len(newNickname) < 3 {
		embed := util.CreateSimpleEmbed(ctx)
		embed.Description = "Nick jest zbyt krótki :worried:"
		_, _ = ctx.Session.ChannelMessageSendEmbed(ctx.ChannelId, embed)
		return errors.SilentNoSellError
	}
	if len(newNickname) > 32 {
		embed := util.CreateSimpleEmbed(ctx)
		embed.Description = "Nick jest zbyt długi :worried:"
		_, _ = ctx.Session.ChannelMessageSendEmbed(ctx.ChannelId, embed)
		return errors.SilentNoSellError
	}
	if strings.ToLower(newNickname) == "young multi" {
		embed := util.CreateSimpleEmbed(ctx)
		embed.Description = "To nielegalne :worried:"
		_, _ = ctx.Session.ChannelMessageSendEmbed(ctx.ChannelId, embed)
		return errors.SilentNoSellError
	}
	if newNickname != newNicknameCleared {
		embed := util.CreateSimpleEmbed(ctx)
		embed.Description = "Nieprawidłowy nick. Spróbuj: `" + newNicknameCleared + "`"
		_, _ = ctx.Session.ChannelMessageSendEmbed(ctx.ChannelId, embed)
		return errors.SilentNoSellError
	}
	err := ctx.Session.GuildMemberNickname(ctx.GuildId, ctx.UserId, newNicknameCleared)
	if err != nil {
		embed := util.CreateSimpleEmbed(ctx)
		embed.Description = "Nie mogę zmienić Ci nicku, poinformuj administracje :worried:"
		_, _ = ctx.Session.ChannelMessageSendEmbed(ctx.ChannelId, embed)
		return errors.SilentNoSellError
	}
	embed := util.CreateSimpleEmbed(ctx)
	embed.Description = "Zakup potwierdzony.\n"
	_, _ = ctx.Session.ChannelMessageSendEmbed(ctx.ChannelId, embed)
	return nil
}
