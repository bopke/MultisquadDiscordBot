package commands

import (
	"github.com/bopke/MultisquadDiscordBot/commands/admin"
	"github.com/bopke/MultisquadDiscordBot/commands/color"
	errors2 "github.com/bopke/MultisquadDiscordBot/commands/errors"
	"github.com/bopke/MultisquadDiscordBot/commands/ratelimits"
	"github.com/bopke/MultisquadDiscordBot/commands/shop"
	"github.com/bopke/MultisquadDiscordBot/commands/vip"
	"github.com/bopke/MultisquadDiscordBot/context"
	"github.com/bopke/MultisquadDiscordBot/database"
	"github.com/bopke/MultisquadDiscordBot/util"
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
)

const prefix = "!"

var (
	commands = map[string]func(*context.Context, []string) (string, error){}
)

func RegisterCommand(command string, handler func(*context.Context, []string) (string, error)) {
	commands[command] = handler
}

func ExecuteCommand(command string, ctx *context.Context, args []string) error {
	handler, ok := commands[command]
	if !ok {
		return errors2.NoSuchCommandError
	}
	if ratelimits.IsTooEarlyToExecute(ctx, command) {
		return nil
	}
	s, err := handler(ctx, args[1:])
	if err == nil {
		return nil
	}
	if err == errors2.IncorrectUsageError {
		embed := util.CreateSimpleEmbed(ctx)
		embed.Title = "Nieprawidłowe użycie"
		embed.Description = s
		_, _ = ctx.Session.ChannelMessageSendEmbed(ctx.ChannelId, embed)
		return nil
	}
	if err == errors2.NoPermissionError {
		embed := util.CreateSimpleEmbed(ctx)
		embed.Title = "Komenda tylko dla admina!"
		_, _ = ctx.Session.ChannelMessageSendEmbed(ctx.ChannelId, embed)
		return nil
	}
	if err == database.DatabaseError {
		embed := util.CreateSimpleEmbed(ctx)
		embed.Title = "Błąd połączenia z bazą danych."
		embed.Description = "Spróbuj ponownie później."
		_, _ = ctx.Session.ChannelMessageSendComplex(ctx.ChannelId, &discordgo.MessageSend{
			Content: "Potrzebne wsparcie: <@205745502266851329>",
			Embed:   embed,
		})
		return nil
	}
	if err == errors2.UnknownError {
		embed := util.CreateSimpleEmbed(ctx)
		embed.Title = "Nieznany błąd."
		embed.Description = "Spróbuj ponownie później.\n" + s
		_, _ = ctx.Session.ChannelMessageSendComplex(ctx.ChannelId, &discordgo.MessageSend{
			Content: "Potrzebne wsparcie: <@205745502266851329>",
			Embed:   embed,
		})
		return nil
	}
	return err
}

func Listener(session *discordgo.Session, event *discordgo.MessageCreate) {
	if len(event.Content) <= len(prefix) {
		return
	}
	if !strings.HasPrefix(event.Content, prefix) {
		return
	}
	event.Content = event.Content[len(prefix):]
	rawArgs := strings.Split(event.Content, " ")
	// TODO: intelligent argument parse
	args := rawArgs
	ctx := context.FromMessageCreate(session, event)
	_ = ctx.FillMember()
	err := ExecuteCommand(strings.ToLower(args[0]), ctx, args)
	if err != nil {
		log.Println("Error while executing command "+event.Content, err)
	}
}

func Init() {
	RegisterCommand("debug", admin.DebugCommand)
	RegisterCommand("kolor", color.ColorCommand)
	RegisterCommand("vip", vip.VipCommand)
	RegisterCommand("sklep", shop.ShopCommand)
	RegisterCommand("kup", shop.BuyCommand)
}
