package vip

import (
	"github.com/bopke/MultisquadDiscordBot/commands/errors"
	"github.com/bopke/MultisquadDiscordBot/context"
	"github.com/bopke/MultisquadDiscordBot/database"
	"github.com/bopke/MultisquadDiscordBot/util"
	"github.com/bopke/MultisquadDiscordBot/vip"
	"github.com/bwmarrin/discordgo"
	"strconv"
)

func VipCommand(ctx *context.Context, args []string) (string, error) {
	if !(util.HasPermission(ctx, discordgo.PermissionAdministrator) || util.HasPermittedRole(ctx)) {
		return "", errors.NoPermissionError
	}
	days := 30
	if len(args) < 1 || len(args) > 2 || !util.IsMention(args[0]) || len(ctx.Message.Mentions) != 1 {
		return "Poprawne użycie: !vip <mention> [ilość dni]", errors.IncorrectUsageError
	}
	if len(args) == 2 {
		num, err := strconv.Atoi(args[1])
		if err != nil {
			return "Ilość dni musi być liczbą!", errors.IncorrectUsageError
		}
		days = num
		if days < 1 {
			return "Ilość dni musi być większa od zera!", errors.IncorrectUsageError
		}
	}
	err := vip.SetVip(ctx.Session, ctx.Message.Mentions[0].ID, days)
	if err == database.DatabaseError {
		return "", err
	} else if err != nil {
		return err.Error(), errors.UnknownError
	}
	return "", nil
}
