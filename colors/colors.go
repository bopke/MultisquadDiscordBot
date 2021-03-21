package colors

import (
	"database/sql"
	"errors"
	"github.com/bopke/MultisquadDiscordBot/config"
	"github.com/bopke/MultisquadDiscordBot/context"
	"github.com/bopke/MultisquadDiscordBot/database"
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
	"time"
)

var (
	NoSuchColorError = errors.New("no such color")

	colors = map[string]string{
		"szary":         "691105469569433632",
		"zielony":       "691099249374527569",
		"mietowy":       "781285965876035614",
		"miętowy":       "781285965876035614",
		"różowy":        "691095736762236939",
		"rozowy":        "691095736762236939",
		"łososiowy":     "781281341705682965",
		"lososiowy":     "781281341705682965",
		"fioletowy":     "691097108148912129",
		"lawendowy":     "781281679976431646",
		"pomarańczowy":  "691095348051050496",
		"pomaranczowy":  "691095348051050496",
		"koralowy":      "781281958364971079",
		"brzoskwiniowy": "691109124230086746",
		"brązowy":       "691097714557190155",
		"brazowy":       "691097714557190155",
		"bordowy":       "691098639732572181",
		"aqua":          "691104242869731330",
		"niebieski":     "691096767118442578",
		"kobaltowy":     "781283711253086278",
		"beżowy":        "781284325282807829",
		"bezowy":        "781284325282807829",
		"khaki":         "781284424281096234",
		"baby pink":     "781290130966708274",
		"baby blue":     "781288898587656203",
	}
)

func getColorExpiredNotification(userId string) string {
	return strings.ReplaceAll("Kolor użytkownika <@{USER_ID}> wygasł.", "{USER_ID}", userId)
}

func getColorNearExpirationNotification(userId string) string {
	return strings.ReplaceAll("Hej, <@{USER_ID}>, za 3 dni wygasa Twój kolor! :worried:", "{USER_ID}", userId)
}

func getColorElongatedNotification(userId string) string {
	return strings.ReplaceAll("<a:drama:595421617354702868> Użytkownik <@{USER_ID}> przedłużył kolor! <a:drama:595421617354702868>", "{USER_ID}", userId)
}

func getColorBoughtNotification(userId string) string {
	return strings.ReplaceAll("<a:drama:595421617354702868> Użytkownik <@{USER_ID}> kupił kolor! <a:drama:595421617354702868>", "{USER_ID}", userId)
}

func GetColorsMentions() (ret []string) {
	contains := func(str string) bool {
		for _, elem := range ret {
			if elem == str {
				return true
			}
		}
		return false
	}
	for _, roleId := range colors {
		if !contains("<@&" + roleId + ">") {
			ret = append(ret, "<@&"+roleId+">")
		}
	}
	return
}

func getColorRoleId(name string) string {
	id, ok := colors[strings.ToLower(name)]
	if !ok {
		return ""
	}
	return id
}

func isColorRole(roleId string) bool {
	for _, colorRoleId := range colors {
		if colorRoleId == roleId {
			return true
		}
	}
	return false
}

func updateUserColorRoles(session *discordgo.Session, userId, colorId string) error {
	member, err := session.GuildMember(config.GuildId, userId)
	if err != nil {
		return err
	}
	hasRole := false
	for _, role := range member.Roles {
		if role == colorId {
			hasRole = true
			continue
		}
		if isColorRole(role) {
			err = session.GuildMemberRoleRemove(config.GuildId, userId, role)
			if err != nil {
				log.Println("updateUserColorRoles Unable to removel role", err)
			}
		}
	}
	if hasRole {
		return nil
	}
	if colorId != "" {
		err = session.GuildMemberRoleAdd(config.GuildId, userId, colorId)
	}
	return err
}

func setColorToNewUser(session *discordgo.Session, userId, colorName, colorId string, days int) error {
	coloredUser := &database.ColoredUser{
		DiscordId:          userId,
		Color:              colorName,
		Valid:              true,
		RoleId:             colorId,
		ExpirationDate:     time.Now().Add(time.Duration(days*24) * time.Hour),
		NotifiedExpiration: false,
	}
	err := database.DbMap.Insert(coloredUser)
	if err != nil {
		return database.DatabaseError
	}
	_, _ = session.ChannelMessageSend(config.AnnouncementChannelId, getColorBoughtNotification(userId))
	return nil
}

func renewUserColor(session *discordgo.Session, coloredUser *database.ColoredUser, colorName, colorId string, days int) error {
	if coloredUser.ExpirationDate.After(time.Now()) {
		coloredUser.ExpirationDate = coloredUser.ExpirationDate.Add(time.Duration(days*24) * time.Hour)
	} else {
		coloredUser.ExpirationDate = time.Now().Add(time.Duration(days*24) * time.Hour)
	}
	coloredUser.NotifiedExpiration = false
	coloredUser.Valid = true
	coloredUser.RoleId = colorId
	coloredUser.Color = colorName
	_, err := database.DbMap.Update(coloredUser)
	if err != nil {
		return database.DatabaseError
	}
	_, _ = session.ChannelMessageSend(config.AnnouncementChannelId, getColorElongatedNotification(coloredUser.DiscordId))
	return nil
}

func SetUserColor(session *discordgo.Session, userId, colorName string, days int) error {
	colorId := getColorRoleId(colorName)
	if colorId == "" {
		return NoSuchColorError
	}
	var coloredUser database.ColoredUser
	err := database.DbMap.SelectOne(&coloredUser, "SELECT * FROM ColoredUsers WHERE discord_id = ?", userId)
	if err == sql.ErrNoRows {
		err = setColorToNewUser(session, userId, colorName, colorId, days)
	} else {
		err = renewUserColor(session, &coloredUser, colorName, colorId, days)
	}
	if err != nil {
		return err
	}
	return updateUserColorRoles(session, userId, colorId)
}

func CheckUserColors(ctx *context.Context) error {
	var coloredUsers []database.ColoredUser
	transaction, err := database.DbMap.Begin()
	if err != nil {
		return database.DatabaseError
	}

	_, err = transaction.Select(&coloredUsers, "SELECT * FROM ColoredUsers")
	if err != nil {
		return database.DatabaseError
	}
	for index, user := range coloredUsers {
		if user.Valid == false {
			continue
		}
		member, err := ctx.Session.GuildMember(config.GuildId, user.DiscordId)
		if err != nil {
			log.Println("checkColors Member loading error", err)
			continue
		}
		if user.ExpirationDate.Before(time.Now()) {
			_, _ = ctx.Session.ChannelMessageSend(config.AnnouncementChannelId, getColorExpiredNotification(user.DiscordId))
			user.Valid = false
			coloredUsers[index].Valid = false
			err = ctx.Session.GuildMemberRoleRemove(config.GuildId, member.User.ID, user.RoleId)
			if err != nil {
				log.Println("checkColors Unable to remove member role", err)
			}
			_, err = transaction.Update(&user)
			if err != nil {
				_ = transaction.Commit()
				return database.DatabaseError
			}
		} else if user.ExpirationDate.Before(time.Now().Add(3 * time.Hour * 24)) {
			if !user.NotifiedExpiration {
				user.NotifiedExpiration = true
				_, _ = transaction.Update(&user)
				_, _ = ctx.Session.ChannelMessageSend(config.AnnouncementChannelId, getColorNearExpirationNotification(user.DiscordId))
			}
		}
	}

	userPermittedColorRole := func(userId string) string {
		for _, coloredUser := range coloredUsers {
			if coloredUser.DiscordId == userId {
				if coloredUser.Valid == false {
					return ""
				}
				return coloredUser.RoleId
			}
		}
		return ""
	}
	err = transaction.Commit()
	if err != nil {
		return database.DatabaseError
	}

	members, err := ctx.Session.GuildMembers(config.GuildId, "0", 1000)
	if err != nil {
		log.Println("checkColors Members loading error", err)
		return err
	}
	for {
		for _, member := range members {
			err = updateUserColorRoles(ctx.Session, member.User.ID, userPermittedColorRole(member.User.ID))
			if err != nil {
				log.Println("checkColors Updating member roles error", err)
			}
		}
		if len(members) != 1000 {
			break
		}
		members, err = ctx.Session.GuildMembers(config.GuildId, members[999].User.ID, 1000)
		if err != nil {
			log.Println(" checkColors Members loading error", err)
			return err
		}
	}
	return nil
}
