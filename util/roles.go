package util

import (
	"errors"
	"github.com/bopke/MultisquadDiscordBot/config"
	"github.com/bopke/MultisquadDiscordBot/context"
	"github.com/bwmarrin/discordgo"
)

var (
	NoSuchRoleError = errors.New("no such role available")
)

func GetRoleID(ctx *context.Context, roleName string) (string, error) {
	guild, err := ctx.Session.Guild(ctx.GuildId)
	if err != nil {
		return "", err
	}
	roles := guild.Roles
	for _, role := range roles {
		if role.Name == roleName {
			return role.ID, nil
		}
	}
	return "", NoSuchRoleError
}

func GetRole(ctx *context.Context, roleName string) (*discordgo.Role, error) {
	guild, err := ctx.Session.Guild(ctx.GuildId)
	if err != nil {
		return nil, err
	}
	roles := guild.Roles
	for _, role := range roles {
		if role.Name == roleName {
			return role, nil
		}
	}
	return nil, NoSuchRoleError
}

func HasRole(ctx *context.Context, roleName string) bool {
	adminRole, err := GetRoleID(ctx, roleName)
	if err != nil {
		return false
	}
	for _, role := range ctx.Member.Roles {
		if role == adminRole {
			return true
		}
	}
	return false
}

func HasRoleId(ctx *context.Context, roleId string) bool {
	for _, role := range ctx.Member.Roles {
		if role == roleId {
			return true
		}
	}
	return false
}

func HasPermission(ctx *context.Context, permission int) bool {
	for _, roleID := range ctx.Member.Roles {
		role, err := ctx.Session.State.Role(ctx.GuildId, roleID)
		if err != nil {
			return false
		}
		if role.Permissions&permission != 0 {
			return true
		}
	}
	return false
}

func HasPermittedRole(ctx *context.Context) bool {
	for _, role := range config.PermittedRolesId {
		if HasRoleId(ctx, role) {
			return true
		}
	}
	return false
}
