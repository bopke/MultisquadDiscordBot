package main

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"log"
)

//funkcja pomocnicza, zbiera ID roli/rangi discorda która nas interesuje
func getRoleID(guildID string, roleName string) (string, error) {
	guild, err := session.Guild(guildID)
	if err != nil {
		log.Println("getRoleID session.Guild(" + guildID + ") " + err.Error())
		return "", err
	}
	roles := guild.Roles
	for _, role := range roles {
		if role.Name == roleName {
			return role.ID, nil
		}
	}
	return "", errors.New("no " + roleName + " role available")
}

//sprawdza, czy dany użytkownik discorda ma taką rangę na tym serwerze.
func hasRole(member *discordgo.Member, roleName, guildID string) bool {
	//z jakiegos powodu w strukturze member GuildID jest puste...
	adminRole, err := getRoleID(guildID, roleName)
	if err != nil {
		log.Println("hasRole getRoleID(" + guildID + ", " + roleName + ") " + err.Error())
		return false
	}
	for _, role := range member.Roles {
		if role == adminRole {
			return true
		}
	}
	return false
}

func hasPermission(member *discordgo.Member, guildID string, permission int) bool {
	for _, roleID := range member.Roles {
		role, err := session.State.Role(guildID, roleID)
		if err != nil {
			log.Println("hasPermisson session.State.Role(" + guildID + ", " + roleID + ") " + err.Error())
			return false
		}
		if role.Permissions&permission != 0 {
			return true
		}
	}
	return false
}
