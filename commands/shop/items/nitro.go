package items

import (
	"github.com/bopke/MultisquadDiscordBot/context"
	"github.com/bopke/MultisquadDiscordBot/util"
)

const nitroLogChannel = "597216492123324442"

func NitroHandler(ctx *context.Context, args []string) error {
	embed := util.CreateSimpleEmbed(ctx)
	embed.Description = "Zakup potwierdzony.\nPowiadomiłem admina. Nitro zostanie wysłane do **48h** w prywatnej wiadomości.\nWłącz **PW** i czekaj na wiadomość."
	_, _ = ctx.Session.ChannelMessageSendEmbed(ctx.ChannelId, embed)
	_, _ = ctx.Session.ChannelMessageSend(nitroLogChannel, "<@"+ctx.UserId+"> zakupił nitro!  <@&598140510082826270> dajcie mu")
	return nil
}
