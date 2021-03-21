package color

import (
	"github.com/bopke/MultisquadDiscordBot/colors"
	"github.com/bopke/MultisquadDiscordBot/commands/errors"
	"github.com/bopke/MultisquadDiscordBot/context"
	"github.com/bopke/MultisquadDiscordBot/database"
	"github.com/bopke/MultisquadDiscordBot/util"
	"github.com/bwmarrin/discordgo"
	"strconv"
)

func ColorCommand(ctx *context.Context, args []string) (string, error) {
	if !(util.HasPermission(ctx, discordgo.PermissionAdministrator) || util.HasPermittedRole(ctx)) {
		return "", errors.NoPermissionError
	}
	days := 30
	if len(args) < 1 || len(args) > 3 || !util.IsMention(args[0]) || len(ctx.Message.Mentions) != 1 {
		return "Poprawne użycie: !color <mention> <nazwa koloru> [ilość dni]", errors.IncorrectUsageError
	}
	if len(args) == 3 {
		num, err := strconv.Atoi(args[2])
		if err != nil {
			return "Ilość dni musi być liczbą!", errors.IncorrectUsageError
		}
		days = num
		if days < 1 {
			return "Ilość dni musi być większa od zera!", errors.IncorrectUsageError
		}
	}
	err := colors.SetUserColor(ctx.Session, ctx.Message.Mentions[0].ID, args[1], days)
	if err == database.DatabaseError {
		return "", err
	} else if err == colors.NoSuchColorError {
		return "Nie ma takiego koloru!", errors.IncorrectUsageError
	} else if err != nil {
		return err.Error(), errors.UnknownError
	}
	return "", nil
}
