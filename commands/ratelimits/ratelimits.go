package ratelimits

import (
	"github.com/bwmarrin/discordgo"
	"strconv"
	"strings"
	"time"

	"github.com/bopke/MultisquadDiscordBot/context"
)

// TODO: REWORK
const (
	RateLimitExceededMessage = "<:nie:586340289497661459> Mordo zwolnij trochę! Odczekaj {SECONDS} sekund i spróbuj ponownie!"
)

var rateLimits map[string]map[string]time.Time = nil

var rateLimitsCommandTime map[string]int = nil

func InitRateLimits(command string, seconds int) {
	if rateLimits == nil {
		rateLimits = map[string]map[string]time.Time{}
		rateLimitsCommandTime = map[string]int{}
	}
	rateLimits[command] = map[string]time.Time{}
	rateLimitsCommandTime[command] = seconds
}

func createRateLimitExceededMessage(duration int) string {
	return strings.ReplaceAll(RateLimitExceededMessage, "{SECONDS}", strconv.Itoa(duration))
}

func tickRatelimitCounter(ctx *context.Context, message *discordgo.Message, duration int) {
	for ; duration > 0; duration-- {
		start := time.Now()
		_, _ = ctx.Session.ChannelMessageEdit(message.ChannelID, message.ID, createRateLimitExceededMessage(duration))
		time.Sleep(start.Add(time.Second).Sub(time.Now()))
	}
	_ = ctx.Session.ChannelMessageDelete(message.ChannelID, message.ID)
}

func IsTooEarlyToExecute(ctx *context.Context, command string) bool {
	commandRateLimits, ok := rateLimits[command]
	if !ok {
		return false
	}
	if commandRateLimits[ctx.UserId].Add(time.Duration(rateLimitsCommandTime[command]) * time.Second).After(time.Now()) {
		timeDifference := rateLimits[command][ctx.UserId].Add(10 * time.Second).Sub(time.Now())
		timeDiff := int(timeDifference.Seconds()) + 1
		msg, err := ctx.Session.ChannelMessageSend(ctx.ChannelId, createRateLimitExceededMessage(timeDiff))
		if err == nil {
			go tickRatelimitCounter(ctx, msg, timeDiff-1)
		}
		return true
	}
	rateLimits[command][ctx.UserId] = time.Now()
	return false
}
