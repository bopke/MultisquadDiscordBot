package nicks

import (
	"github.com/bopke/MultisquadDiscordBot/config"
	"github.com/bwmarrin/discordgo"
	"log"
)

const allowedNicknameChars = "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM1234567890_ ęóąśłżźćńĘÓĄŚŁŻŹĆŃ"

func FixNickname(session *discordgo.Session, member *discordgo.Member) {
	var newNickname string
	if member.Nick == "" {
		newNickname = ClearUsername(member.User.Username)
	} else {
		newNickname = ClearUsername(member.Nick)
		if len(newNickname) < 3 {
			newNickname = ClearUsername(member.User.Username)
		}
	}
	if len(newNickname) < 3 {
		newNickname = ClearUsername("Zmień nick")
	}
	if member.Nick == newNickname || (newNickname == member.User.Username && (member.Nick == member.User.Username || member.Nick == "")) {
		return
	}
	log.Println("Changing nickname of " + member.User.Username + "#" + member.User.Discriminator + " to " + newNickname)
	err := session.GuildMemberNickname(config.GuildId, member.User.ID, newNickname)
	if err != nil {
		log.Println("Unable to change nickname of " + member.User.Username + "#" + member.User.Discriminator + ".")
		log.Println(err)
	}

}
func ClearUsername(username string) string {
	cleared := ""
	for _, char := range username {
		for _, legalChar := range allowedNicknameChars {
			if char == legalChar {
				cleared += string(char)
			}
		}
	}
	for ; len(cleared) > 0 && cleared[0] == ' '; cleared = cleared[1:] {
		// if space is allowed, it still cannot be the first character
	}
	for ; len(cleared) > 0 && cleared[len(cleared)-1] == ' '; cleared = cleared[:len(cleared)-1] {
		// if space is allowed, it still cannot be last character
	}
	return cleared
}

func CheckNicknames(session *discordgo.Session) error {
	after := ""
	for {
		members, err := session.GuildMembers(config.GuildId, after, 1000)
		if err != nil {
			return err
		}
		for _, member := range members {
			FixNickname(session, member)
		}
		if len(members) == 1000 {
			after = members[999].User.ID
		} else {
			break
		}
	}
	return nil
}
