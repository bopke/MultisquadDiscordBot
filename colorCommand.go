package main

import (
	"database/sql"
	"github.com/bwmarrin/discordgo"
	"log"
	"strconv"
	"strings"
	"time"
)

func checkIfValidColorCode(s string) bool {
	// 2020 temporary workaround
	// i dont want to see this stuff in 2021 and so on
	// but I probably will
	switch strings.ToLower(s) {
	case "niebieski":
		fallthrough
	case "aqua":
		fallthrough
	case "bordowy":
		fallthrough
	case "brązowy":
		fallthrough
	case "brzoskwiniowy":
		fallthrough
	case "pomarańczowy":
		fallthrough
	case "fioletowy":
		fallthrough
	case "różowy":
		fallthrough
	case "zielony":
		fallthrough
	case "szary":
		return true
	default:
		return false
	}
	return false
	/*
		if len(s) != 7 {
			return false
		}
		s = strings.ToUpper(s)
		helper := "0123456789ABCDEF"
		if s[0] != '#' {
			return false
		}
		for pos, i := range s {
			if pos == 0 {
				continue
			}
			ok := false
			for _, j := range helper {
				if i == j {
					ok = true
					break
				}
			}
			if ok == false {
				return false
			}
		}
		return true*/
}

func setColorRole(s *discordgo.Session, message *discordgo.MessageCreate, args []string) error {
	roleId, err := getRoleID(message.GuildID, args[2])
	if err != nil {
		log.Println("Nie udalo sie pobrac roli koloru " + err.Error())
		role, err := session.GuildRoleCreate(message.GuildID)
		if err != nil {
			return err
		}
		roleId = role.ID
		color, err := strconv.ParseUint(args[2][1:], 16, 32)
		if err != nil {
			return err
		}
		_, err = s.GuildRoleEdit(message.GuildID, role.ID, args[2], int(color), false, role.Permissions, false)
		if err != nil {
			_ = session.GuildRoleDelete(message.GuildID, roleId)
			msg, err2 := s.ChannelMessageSend(message.ChannelID, Locale.UnexpectedApiError)
			if err2 == nil {
				time.Sleep(20 * time.Second)
				_ = s.ChannelMessageDelete(msg.ChannelID, msg.ID)
			}
			return err
		}
		roles, err := s.GuildRoles(message.GuildID)
		if err != nil {
			msg, err := s.ChannelMessageSend(message.ChannelID, Locale.UnexpectedApiError)
			if err == nil {
				time.Sleep(20 * time.Second)
				_ = s.ChannelMessageDelete(msg.ChannelID, msg.ID)
			}
			return err
		}
		var newRoles []*discordgo.Role
		LimiterRole, _ := getRole(message.GuildID, Config.ColorRoleHierarchyLimiterRole)
		for _, oldRole := range roles {
			if oldRole.Position >= LimiterRole.Position {
				oldRole.Position += 1
			}
			if oldRole.ID == roleId {
				oldRole.Position = LimiterRole.Position
			}
			newRoles = append(newRoles, oldRole)
			if oldRole.Name == Config.ColorRoleHierarchyLimiterRole {
				newRoles = append(newRoles, role)
			}
		}
		_, err = s.GuildRoleReorder(message.GuildID, newRoles)
		if err != nil {
			msg, err := s.ChannelMessageSend(message.ChannelID, Locale.UnexpectedApiError)
			if err == nil {
				time.Sleep(20 * time.Second)
				_ = s.ChannelMessageDelete(msg.ChannelID, msg.ID)
			}
			return err
		}
	}
	err = s.GuildMemberRoleAdd(message.GuildID, message.Mentions[0].ID, roleId)
	return err
}

func handleColorCommand(s *discordgo.Session, message *discordgo.MessageCreate) {
	_ = s.ChannelMessageDelete(message.ChannelID, message.ID)
	member, err := s.GuildMember(message.GuildID, message.Author.ID)
	if err != nil {
		log.Println("Błąd pobierania twórcy wiadomości!\n" + err.Error())
		return
	}
	if !hasPermission(member, message.GuildID, discordgo.PermissionAdministrator) {
		msg, err := s.ChannelMessageSend(message.ChannelID, Locale.NoAdminPermission)
		if err == nil {
			time.Sleep(20 * time.Second)
			_ = s.ChannelMessageDelete(msg.ChannelID, msg.ID)
		}
		return
	}
	// dzielimy wiadomość po spacjach dla wygody
	args := strings.Split(message.Content, " ")
	var length int
	if message.Mentions != nil && len(message.Mentions) == 0 {
		msg, err := s.ChannelMessageSend(message.ChannelID, Locale.ColorIncorrectUser)
		if err == nil {
			time.Sleep(20 * time.Second)
			_ = s.ChannelMessageDelete(msg.ChannelID, msg.ID)
		}
		return
	}
	if len(args) < 3 {
		msg, err := s.ChannelMessageSend(message.ChannelID, Locale.ColorIncorrectColorCode)
		if err == nil {
			time.Sleep(20 * time.Second)
			_ = s.ChannelMessageDelete(msg.ChannelID, msg.ID)
		}
		return
	}
	if len(args) >= 4 {
		length, err = strconv.Atoi(args[3])
		if err != nil {
			log.Println("Błąd argumentu dni \"" + args[3] + "\" " + err.Error())
			msg, err := s.ChannelMessageSend(message.ChannelID, Locale.ColorIncorrectDaysCount)
			if err == nil {
				time.Sleep(20 * time.Second)
				_ = s.ChannelMessageDelete(msg.ChannelID, msg.ID)
			}
			return
		}
	} else {
		length = Config.ColorDefaultLength
	}
	if !checkIfValidColorCode(args[2]) {
		msg, err := s.ChannelMessageSend(message.ChannelID, Locale.ColorIncorrectColorCode)
		if err == nil {
			time.Sleep(20 * time.Second)
			_ = s.ChannelMessageDelete(msg.ChannelID, msg.ID)
		}
		return
	}
	var coloredUser ColoredUser
	err = DbMap.SelectOne(&coloredUser, "SELECT * FROM ColoredUsers WHERE discord_id = ?", message.Mentions[0].ID)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Błąd połączenia z bazą danych!\n" + err.Error())
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.DatabaseError)
		return
	}
	if err == sql.ErrNoRows {
		coloredUser.DiscordID = message.Mentions[0].ID
		coloredUser.Valid = true
		coloredUser.ExpirationDate = time.Now().Add((time.Hour * 24) * time.Duration(length))
		coloredUser.NotifiedExpiration = false
		coloredUser.Color = args[2]
		coloredUser.RoleId, _ = getRoleID(message.GuildID, args[2])
		err = DbMap.Insert(&coloredUser)
		if err != nil {
			log.Println("Błąd połączenia z bazą danych!\n" + err.Error())
			_, _ = s.ChannelMessageSend(message.ChannelID, Locale.DatabaseError)
			return
		}
		err = setColorRole(s, message, args)
		if err != nil {
			log.Println("Błąd przydzielania rangi!\n" + err.Error())
			msg, err := s.ChannelMessageSend(message.ChannelID, Locale.UnexpectedApiError)
			if err == nil {
				time.Sleep(20 * time.Second)
				_ = s.ChannelMessageDelete(msg.ChannelID, msg.ID)
			}
			return
		}
		log.Println("Kolor użytkownika " + message.Mentions[0].Username + "#" + message.Mentions[0].Discriminator + " został utworzony na " + strconv.Itoa(length) + " dni")
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.ColorInserted)
		_, _ = s.ChannelMessageSend(Config.AnnouncementChannelId, strings.Replace(Locale.ColorAnnouncementInserted, "{MENTION}", message.Mentions[0].Mention(), -1))
		return
	}
	coloredUser.Color = args[2]
	coloredUser.RoleId, _ = getRoleID(message.GuildID, args[2])
	coloredUser.Valid = true
	if coloredUser.ExpirationDate.Before(time.Now()) {
		coloredUser.ExpirationDate = time.Now().Add((time.Hour * 24) * time.Duration(length))
	} else {
		coloredUser.ExpirationDate = coloredUser.ExpirationDate.Add((time.Hour * 24) * time.Duration(length))
	}
	coloredUser.NotifiedExpiration = false
	_, err = DbMap.Update(&coloredUser)
	if err != nil {
		log.Println("Błąd połączenia z bazą danych!\n" + err.Error())
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.DatabaseError)
		return
	}
	err = setColorRole(s, message, args)
	if err != nil {
		log.Println("Błąd nadawania rangi " + err.Error())
		msg, err := s.ChannelMessageSend(message.ChannelID, Locale.UnexpectedApiError)
		if err == nil {
			time.Sleep(20 * time.Second)
			_ = s.ChannelMessageDelete(msg.ChannelID, msg.ID)
		}
		return
	}
	log.Println("Kolor użytkownika " + message.Mentions[0].Username + "#" + message.Mentions[0].Discriminator + " został utworzony na " + strconv.Itoa(length) + " dni")
	_, _ = s.ChannelMessageSend(message.ChannelID, Locale.ColorUpdated)
	_, _ = s.ChannelMessageSend(Config.AnnouncementChannelId, strings.Replace(Locale.ColorAnnouncementUpdated, "{MENTION}", message.Mentions[0].Mention(), -1))
}
