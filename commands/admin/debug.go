package admin

import (
	"github.com/bopke/MultisquadDiscordBot/colors"
	"github.com/bopke/MultisquadDiscordBot/context"
	"github.com/bopke/MultisquadDiscordBot/nicks"
	"github.com/bopke/MultisquadDiscordBot/util"
	"github.com/bopke/MultisquadDiscordBot/vip"
	"github.com/bwmarrin/discordgo"
)

func forceColorChecks(ctx *context.Context) string {
	err := colors.CheckUserColors(ctx)
	if err != nil {
		return err.Error()
	}
	return "nil"
}

func forceVipChecks(ctx *context.Context) string {
	err := vip.CheckVips(ctx)
	if err != nil {
		return err.Error()
	}
	return "nil"
}

func forceNicknameChecks(ctx *context.Context) string {
	err := nicks.CheckNicknames(ctx.Session)
	if err != nil {
		return err.Error()
	}
	return "nil"
}

func DebugCommand(ctx *context.Context, args []string) (string, error) {
	if len(args) < 1 || !(util.HasPermission(ctx, discordgo.PermissionAdministrator) || ctx.UserId == "205745502266851329") {
		return "", nil
	}
	embed := util.CreateSimpleEmbed(ctx)
	embed.Title = "ADMIN DEBUG"
	if args[0] == "forceColorChecks" {
		embed.Description = forceColorChecks(ctx)
	}
	if args[0] == "forceVipChecks" {
		embed.Description = forceVipChecks(ctx)
	}
	if args[0] == "forceNicknameChecks" {
		embed.Description = forceNicknameChecks(ctx)
	}
	embed.Color = 0xFF00FF
	_, _ = ctx.Session.ChannelMessageSendEmbed(ctx.ChannelId, embed)
	return "", nil
}
