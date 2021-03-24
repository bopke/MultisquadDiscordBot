package shop

import (
	"github.com/bopke/MultisquadDiscordBot/colors"
	"github.com/bopke/MultisquadDiscordBot/context"
	"github.com/bopke/MultisquadDiscordBot/util"
	"github.com/bwmarrin/discordgo"
	"strings"
)

func ShopCommand(ctx *context.Context, args []string) (string, error) {
	embed := util.CreateSimpleEmbed(ctx)
	embed.Description = "Kup przedmiot za pomocą komendy `!kup`"
	embed.Author = &discordgo.MessageEmbedAuthor{
		Name:         ctx.Message.Author.Username + "#" + ctx.Message.Author.Discriminator,
		IconURL:      ctx.Message.Author.AvatarURL("128"),
		ProxyIconURL: "",
	}
	embed.Fields = []*discordgo.MessageEmbedField{
		{Name: "1. NICK - 500 <a:moneta:613020692346175628>",
			Value:  "Zmiana nicku.\nKup poleceniem `!kup nick <nowy nick>`, np `!kup nick Young Multi`.",
			Inline: false,
		},
		{Name: "2. KOLOR - 15.000 <a:moneta:613020692346175628>",
			Value:  "Zmiana koloru nicku.\nKup poleceniem `!kup kolor <nazwa koloru>`, np `!kup kolor pomaranczowy`.\nDostępne kolory:\n" + strings.Join(colors.GetColorsMentions(), "\n"),
			Inline: false,
		},
		/*{Name: "3. MYSTERYBOX - 5.000 <a:moneta:613020692346175628>",
			Value:  "Losowanie itemka (czymkolwiek to jest).\nKup poleceniem `!kup mysterybox`.",
			Inline: false,
		},*/
		{Name: "3. VIP - 65.000 <a:moneta:613020692346175628>",
			Value:  "VIP na 30 dni.\nKup poleceniem `!kup vip`.",
			Inline: false,
		},
		{Name: "4. NITRO - 90.000 <a:moneta:613020692346175628>",
			Value:  "Nitro na 30 dni.\nWysyłane ręcznie przez administratora. Kup poleceniem `!kup nitro`.",
			Inline: false,
		},
		{Name: "5. FLEXER - 100.000 <a:moneta:613020692346175628>",
			Value:  "Ranga Flexer.\nKup poleceniem `!kup flexer`.",
			Inline: false,
		},
		{
			Name:   "6. Odznaka sezonowa - 50.000 <a:moneta:613020692346175628>",
			Value:  "Odznaka sezonowa.\nKup poleceniem `!kup sezon`.",
			Inline: false,
		},
	}
	_, _ = ctx.Session.ChannelMessageSendEmbed(ctx.ChannelId, embed)
	return "", nil
}
