package vip

import (
	"database/sql"
	"github.com/bopke/MultisquadDiscordBot/config"
	"github.com/bopke/MultisquadDiscordBot/context"
	"github.com/bopke/MultisquadDiscordBot/database"
	"github.com/bopke/MultisquadDiscordBot/util"
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
	"time"
)

const (
	vipRoleId = "579717933736132620"
)

func getVipExpiredNotification(userId string) string {
	return strings.ReplaceAll("Vip użytkownika <@{USER_ID}> wygasł.", "{USER_ID}", userId)
}

func getVipNearExpirationNotification(userId string) string {
	return strings.ReplaceAll("Hej, <@{USER_ID}>, za 3 dni wygasa Twój vip! :worried:", "{USER_ID}", userId)
}

func getVipElongatedNotification(userId string) string {
	return strings.ReplaceAll("<a:drama:595421617354702868> Użytkownik <@{USER_ID}> przedłużył vipa! <a:drama:595421617354702868>", "{USER_ID}", userId)
}

func getVipBoughtNotification(userId string) string {
	return strings.ReplaceAll("<a:drama:595421617354702868> Użytkownik <@{USER_ID}> kupił vipa! <a:drama:595421617354702868>", "{USER_ID}", userId)
}

func updateVipRole(session *discordgo.Session, userId string, isVip bool) error {
	member, err := session.GuildMember(config.GuildId, userId)
	if err != nil {
		return err
	}
	if util.HasRoleId(&context.Context{
		Session: session,
		Member:  member,
		UserId:  userId,
	}, vipRoleId) {
		if !isVip {
			err = session.GuildMemberRoleRemove(config.GuildId, userId, vipRoleId)
			if err != nil {
				log.Println("updateVipRole Unable to remove role", err)
			}
		}
	} else {
		if isVip {
			err = session.GuildMemberRoleAdd(config.GuildId, userId, vipRoleId)
			if err != nil {
				log.Println("updateVipRole Unable to remove role", err)
			}
		}
	}
	return err
}

func setVipToNewUser(session *discordgo.Session, userId string, days int) error {
	linkedUser := &database.LinkedUsers{
		DiscordId:          userId,
		Valid:              true,
		ExpirationDate:     time.Now().Add(time.Duration(days*24) * time.Hour),
		NotifiedExpiration: false,
	}
	err := database.DbMap.Insert(linkedUser)
	if err != nil {
		return database.DatabaseError
	}
	_, _ = session.ChannelMessageSend(config.AnnouncementChannelId, getVipBoughtNotification(userId))
	return nil
}

func renewVip(session *discordgo.Session, linkedUser *database.LinkedUsers, days int) error {
	if linkedUser.ExpirationDate.After(time.Now()) {
		linkedUser.ExpirationDate = linkedUser.ExpirationDate.Add(time.Duration(days*24) * time.Hour)
	} else {
		linkedUser.ExpirationDate = time.Now().Add(time.Duration(days*24) * time.Hour)
	}
	linkedUser.NotifiedExpiration = false
	linkedUser.Valid = true
	_, err := database.DbMap.Update(linkedUser)
	if err != nil {
		return database.DatabaseError
	}
	_, _ = session.ChannelMessageSend(config.AnnouncementChannelId, getVipElongatedNotification(linkedUser.DiscordId))
	return nil
}

func SetVip(session *discordgo.Session, userId string, days int) error {
	var linkedUser database.LinkedUsers
	err := database.DbMap.SelectOne(&linkedUser, "SELECT * FROM LinkedUsers WHERE discord_id = ?", userId)
	if err == sql.ErrNoRows {
		err = setVipToNewUser(session, userId, days)
	} else {
		err = renewVip(session, &linkedUser, days)
	}
	if err != nil {
		return err
	}
	return updateVipRole(session, userId, true)
}

func CheckVips(ctx *context.Context) error {
	var linkedUsers []database.LinkedUsers
	transaction, err := database.DbMap.Begin()
	if err != nil {
		return database.DatabaseError
	}

	_, err = transaction.Select(&linkedUsers, "SELECT * FROM LinkedUsers")
	if err != nil {
		return database.DatabaseError
	}
	for index, user := range linkedUsers {
		if user.Valid == false {
			continue
		}
		member, err := ctx.Session.GuildMember(config.GuildId, user.DiscordId)
		if err != nil {
			log.Println("checkVips Member loading error", err)
			continue
		}
		if user.ExpirationDate.Before(time.Now()) {
			_, _ = ctx.Session.ChannelMessageSend(config.AnnouncementChannelId, getVipExpiredNotification(user.DiscordId))
			user.Valid = false
			linkedUsers[index].Valid = false
			err = ctx.Session.GuildMemberRoleRemove(config.GuildId, member.User.ID, vipRoleId)
			if err != nil {
				log.Println("checkVips Unable to remove member role", err)
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
				_, _ = ctx.Session.ChannelMessageSend(config.AnnouncementChannelId, getVipNearExpirationNotification(user.DiscordId))
			}
		}
	}

	shouldUserBeVip := func(userId string) bool {
		var linkedUser database.LinkedUsers
		err := database.DbMap.SelectOne(&linkedUser, "SELECT * FROM LinkedUsers WHERE discord_id=?", userId)
		if err != nil {
			return false
		}
		if linkedUser.Valid == false {
			return false
		}
		return true
	}
	err = transaction.Commit()
	if err != nil {
		return database.DatabaseError
	}

	members, err := ctx.Session.GuildMembers(config.GuildId, "0", 1000)
	if err != nil {
		log.Println("checkVips Members loading error", err)
		return err
	}
	for {
		for _, member := range members {
			err = updateVipRole(ctx.Session, member.User.ID, shouldUserBeVip(member.User.ID))
			if err != nil {
				log.Println("checkVips Updating member roles error", err)
			}
		}
		if len(members) != 1000 {
			break
		}
		members, err = ctx.Session.GuildMembers(config.GuildId, members[999].User.ID, 1000)
		if err != nil {
			log.Println(" checkVips Members loading error", err)
			return err
		}
	}
	return nil
}
