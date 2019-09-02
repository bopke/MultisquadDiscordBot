package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/robfig/cron"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// sesja połączenia z botem musi być globalna, żeby cykliczne zadanie sprawdzania użytkowników mogło działać
var session *discordgo.Session

//stan dodawania steamId do bazy, używane przy funkcji łączenia użytkownika discorda z kontem steam.
type State byte

const (
	INSERTED State = iota
	UPDATED
	ERROR
)

func main() {
	log.Println("Warming up...")
	Config.load()
	Locale.load()
	InitDB()
	InitRateLimits()
	var err error
	session, err = discordgo.New("Bot " + Config.DiscordToken)
	if err != nil {
		panic(err)
	}
	session.AddHandler(OnMessageCreate)
	session.AddHandler(OnGuildMemberUpdate)
	err = session.Open()
	if err != nil {
		panic(err)
	}

	// cron - narzędzie do cyklicznego wykonywania zadania. Co minutę będzie odpalać funkcję checkUsers.
	c := cron.New()
	_ = c.AddFunc("0 * * * * *", checkUsers)
	c.Start()

	go inits()
	log.Println("Started.")
	// ten kanał powoduje utrzymanie działania programu dopóki nie przyjdzie do niego sygnał od systemu operacyjnego, że pora się zwijać
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	err = session.Close()
	if err != nil {
		panic(err)
	}
}

func inits() {
	checkNicknames("")
}

func checkNicknames(after string) {
	members, err := session.GuildMembers(Config.ServerId, after, 1000)
	if err != nil {
		log.Println("Error checking nicknames " + err.Error())
		return
	}
	for _, member := range members {
		fixNickname(member)
	}
	if len(members) == 1000 {
		checkNicknames(members[999].User.ID)
	}
}

//funkcja odpalana cyklicznie, sprawdza czy wszyscy w bazie nadal są na serwerze i czy są na nim vipami.
func checkUsers() {
	roleId, err := getRoleID(Config.ServerId, Config.PermittedRoleName)
	if err != nil {
		log.Println("Błąd pobierania informacji o roli!\n" + err.Error())
		return
	}
	//zbieramy wszystkie zapisane ID z discorda
	var linkedUsers []LinkedUsers
	_, err = DbMap.Select(&linkedUsers, "SELECT id,discord_id,valid,expiration_date,notified_expiration FROM LinkedUsers WHERE valid=true")
	var hasRole bool
	for _, linkedUser := range linkedUsers {
		member, err := session.GuildMember(Config.ServerId, linkedUser.DiscordID)
		// jeżeli użytkownika nie ma już na serwerze - odcinamy go.
		if member == nil && err != nil && err.Error() == "HTTP 404 Not Found, {\"code\": 10007, \"message\": \"Unknown Member\"}" {
			if linkedUser.Valid == true {
				_, _ = DbMap.Exec("UPDATE LinkedUsers SET valid=false WHERE discord_id=?", linkedUser.DiscordID)
			}
			continue
		}
		if err != nil {
			log.Println("Błąd przy pobieraniu użytkownika!\n" + err.Error())
			continue
		}
		hasRole = false
		for _, role := range member.Roles {
			if role == roleId {
				hasRole = true
				break
			}
		}
		if linkedUser.ExpirationDate.Before(time.Now()) {
			if hasRole {
				log.Println("Wykryłem że użytkownik " + member.User.Username + "#" + member.User.Discriminator + " ma vipa, ale już nie powinien go mieć. Odbieram.")
				err = session.GuildMemberRoleRemove(Config.ServerId, linkedUser.DiscordID, roleId)
				if err != nil {
					log.Println("Błąd usuwania rangi użytkownika " + err.Error())
					continue
				}
				_, _ = session.ChannelMessageSend(Config.AnnouncementChannelId, strings.Replace(Locale.VipExpiredNotification, "{MENTION}", member.Mention(), -1))
				_, _ = DbMap.Exec("UPDATE LinkedUsers SET valid=false WHERE discord_id=?", linkedUser.DiscordID)
				continue
			}
		} else if linkedUser.ExpirationDate.Before(time.Now().Add(3 * time.Hour * 24)) {
			if !hasRole {
				log.Println("Wykryłem że użytkownik " + member.User.Username + "#" + member.User.Discriminator + " nie ma vipa, a powinien go mieć. Nadaję.")
				err = session.GuildMemberRoleAdd(Config.ServerId, linkedUser.DiscordID, roleId)
				if err != nil {
					log.Println("Błąd dodawania rangi użytkownika " + err.Error())
					continue
				}
			}
			if !linkedUser.NotifiedExpiration {
				linkedUser.NotifiedExpiration = true
				_, _ = DbMap.Update(&linkedUser)
				_, _ = session.ChannelMessageSend(Config.AnnouncementChannelId, strings.Replace(Locale.VipNearExpirationNotification, "{MENTION}", member.Mention(), -1))
			}
		}
		if !hasRole {
			log.Println("Wykryłem że użytkownik " + member.User.Username + "#" + member.User.Discriminator + " nie ma vipa, a powinien go mieć. Nadaję.")
			err = session.GuildMemberRoleAdd(Config.ServerId, linkedUser.DiscordID, roleId)
			if err != nil {
				log.Println("Błąd dodawania rangi użytkownika " + err.Error())
				continue
			}
		}
	}

	members, err := session.GuildMembers(Config.ServerId, "0", 1000)
	if err != nil {
		log.Println("Błąd pobierania użytkowników serwera " + err.Error())
		return
	}
	for {
		for _, member := range members {
			hasRole = false
			for _, role := range member.Roles {
				if role == roleId {
					hasRole = true
					break
				}
			}
			if hasRole != true {
				continue
			}
			for _, linkedUser := range linkedUsers {
				if linkedUser.DiscordID == member.User.ID {
					hasRole = false
					break
				}
			}
			if hasRole {
				log.Println("Wykryłem że użytkownik " + member.User.Username + "#" + member.User.Discriminator + " ma vipa, ale już nie powinien go mieć. Odbieram.")
				_ = session.GuildMemberRoleRemove(Config.ServerId, member.User.ID, roleId)
			}
		}
		if len(members) != 1000 {
			break
		}
		members, err = session.GuildMembers(Config.ServerId, members[999].User.ID, 1000)
		if err != nil {
			log.Println("Błąd pobierania użytkowników serwera " + err.Error())
			return
		}
	}
}

func fixNickname(member *discordgo.Member) {
	var newNickname string
	if member.Nick == "" {
		newNickname = clearUsername(member.User.Username)
	} else {
		newNickname = clearUsername(member.Nick)
		if len(newNickname) < 3 {
			newNickname = clearUsername(member.User.Username)
		}
	}
	if len(newNickname) < 3 {
		newNickname = clearUsername("Zmień nick")
	}
	if member.Nick == newNickname || (newNickname == member.User.Username && (member.Nick == member.User.Username || member.Nick == "")) {
		return
	}
	log.Println(newNickname)
	err := session.GuildMemberNickname(Config.ServerId, member.User.ID, newNickname)
	if err != nil {
		log.Println("Unable to change nickname of " + member.User.Username + "#" + member.User.Discriminator + ".")
		log.Println(err)
	}

}
func clearUsername(username string) string {
	cleared := ""
	for _, char := range username {
		for _, legalChar := range Config.AllowedNicknameChars {
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
