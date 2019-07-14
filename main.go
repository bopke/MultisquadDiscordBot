package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	_ "github.com/go-sql-driver/mysql"
	"github.com/robfig/cron"
	"gopkg.in/gorp.v2"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// struktura konfiguracji
type Configuration struct {
	MysqlLogin        string `json:"mysql_login"`
	MysqlPassword     string `json:"mysql_password"`
	MysqlDatabase     string `json:"mysql_database"`
	MysqlHost         string `json:"mysql_host"`
	MysqlPort         int    `json:"mysql_port"`
	CommandName       string `json:"command_name"`
	DiscordToken      string `json:"discord_token"`
	PermittedRoleName string `json:"permitted_role_name"`
	ServerId          string `json:"server_id"`
}

// zmienna globalna która będzie przechowywać konfigurację
var Config = new(Configuration)

// "metoda" struktury konfiguracji umożliwiająca jej załadowanie
func (c *Configuration) load() {
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Panic(err)
	}
	defer configFile.Close()
	// dekodujemy całą zawartość pliku na utworzoną strukturę
	err = json.NewDecoder(configFile).Decode(c)
	if err != nil {
		log.Panic("loadConfig Decoder.Decode(c) " + err.Error())
	}
	return
}

//struktura z tekstami
type Locales struct {
	NoPermission       string `json:"no_permission"`
	InvalidProfileLink string `json:"invalid_profile_link"`
	InvalidProfileId   string `json:"invalid_profile_id"`
	DatabaseError      string `json:"database_error"`
	SteamIdUpdated     string `json:"steamid_updated"`
	SteamIdInserted    string `json:"steamid_inserted"`
	UnexpectedError    string `json:"unexpected_error"`
}

var Locale = new(Locales)

// "metoda" struktury pliku tłumaczeń umożliwiająca jej załadowanie
func (l *Locales) load() {
	localeFile, err := os.Open("locale.json")
	if err != nil {
		log.Panic(err)
	}
	defer localeFile.Close()
	// dekodujemy całą zawartość pliku na utworzoną strukturę
	err = json.NewDecoder(localeFile).Decode(l)
	if err != nil {
		log.Panic("loadLocales Decoder.Decode(l) " + err.Error())
	}
	return
}

// struktura tabeli w bazie danych
type LinkedUsers struct {
	Id             int       `db:"id,primarykey,autoincrement"`
	DiscordID      string    `db:"discord_id,size:255"`
	SteamID64      string    `db:"steam_id,size:255"`
	Valid          bool      `db:"valid"`
	ExpirationDate time.Time `db:"expiration_date"`
}

var DbMap gorp.DbMap

//error używany przy zwracaniu błędu braku profilu w funkcji wydobywania steamid64
var noSuchProfileError = errors.New("no such profile")

//inicjujemy bazę danych
func InitDB() {
	mysqlConnectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", Config.MysqlLogin, Config.MysqlPassword, Config.MysqlHost, Config.MysqlPort, Config.MysqlDatabase)
	db, err := sql.Open("mysql", mysqlConnectionString)
	if err != nil {
		log.Panic("Błąd połączenia z bazą danych!\n" + err.Error())
	}
	// DbMap to schemat bazy danych, dialect oznacza jaką dokładnie baze chcemy używać - mysql, sqlite, postgresql itp...
	DbMap = gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8MB4"}}

	//dodajemy mapie tabelę której struktura jest zdefiniowana wyżej.
	DbMap.AddTableWithName(LinkedUsers{}, "LinkedUsers").SetKeys(true, "id")

	err = DbMap.CreateTablesIfNotExists()
	if err != nil {
		log.Panic("InitDB DbMap.CreateTablesIfNotExists() " + err.Error())
	}
}

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
	var err error
	session, err = discordgo.New("Bot " + Config.DiscordToken)
	if err != nil {
		panic(err)
	}
	session.AddHandler(OnMessageCreate)
	err = session.Open()
	if err != nil {
		panic(err)
	}

	// cron - narzędzie do cyklicznego wykonywania zadania. Co minutę będzie odpalać funkcję checkUsers.
	c := cron.New()
	_ = c.AddFunc("0 * * * * *", checkUsers)
	c.Start()

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

// funkcja ta przyjmuje każdą wiadomość która zostanie wysłana na kanałach, które widzi bot i analizuje ją.
func OnMessageCreate(s *discordgo.Session, message *discordgo.MessageCreate) {
	//jeżeli wiadomość jest na serwerze innym niż nasz oczekiwany to wywalać z tymi komendami.
	if message.GuildID != Config.ServerId {
		return
	}
	// jeżeli wiadomość nie zaczyna się od naszej komendy to nie analizujemy dalej
	if !strings.HasPrefix(message.Content, Config.CommandName) {
		return
	}
	member, err := s.GuildMember(message.GuildID, message.Author.ID)
	if err != nil {
		log.Println("Błąd pobierania twórcy wiadomości!\n" + err.Error())
		return
	}
	if !hasRole(member, Config.PermittedRoleName, message.GuildID) {
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.NoPermission)
		return
	}
	log.Println(message.Author.Username + "#" + message.Author.Discriminator + " wykonał polecenie: " + message.Content)
	// dzielimy wiadomość po spacjach dla wygody
	args := strings.Split(message.Content, " ")
	if len(args) < 2 {
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.InvalidProfileLink)
		return
	}
	// link do profilu może zacząć się od https:// ...
	if strings.HasPrefix(args[1], "https://") {
		args[1] = args[1][8:]
	} else
	// lub od http://...
	if strings.HasPrefix(args[1], "http://") {
		args[1] = args[1][7:]
	}
	//albo po prostu od razu od adresu do strony steama. Pozwólmy ludziom wklejać różne warianty.
	if !strings.HasPrefix(args[1], "steamcommunity.com/id/") {
		if strings.HasPrefix(args[1], "steamcommunity.com/profiles/") {
			// interesuje nas tylko to, co jest po /profiles/
			args[1] = args[1][28:]
			err := validateSteamId(args[1])
			if err == nil {
				state := linkUser(message.Author.ID, args[1])
				if state == ERROR {
					_, _ = s.ChannelMessageSend(message.ChannelID, Locale.DatabaseError)
				} else if state == UPDATED {
					_, _ = s.ChannelMessageSend(message.ChannelID, Locale.SteamIdUpdated)
				} else if state == INSERTED {
					_, _ = s.ChannelMessageSend(message.ChannelID, Locale.SteamIdInserted)
				}
				return
			} else {
				_, _ = s.ChannelMessageSend(message.ChannelID, Locale.UnexpectedError)
				return
			}
		}
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.InvalidProfileLink)
		return
	}
	//interesuje nas tylko to, co jest po id/
	args[1] = args[1][22:]
	// i na podstawie tego możemy pobrać steamID
	steamId, err := getSteamIdForProfileId(args[1])
	if err != nil {
		// ale ktoś może podać zły link i taki profil nie istnieje!
		if err == noSuchProfileError {
			_, _ = s.ChannelMessageSend(message.ChannelID, Locale.InvalidProfileId)
			return
		}
		// może też wystąpić jakiś nieoczekiwany błąd..
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.UnexpectedError)
		log.Println("Błąd przy odczycie danych z API!\n" + err.Error())
		return
	}
	state := linkUser(message.Author.ID, steamId)
	if state == ERROR {
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.DatabaseError)
	} else if state == UPDATED {
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.SteamIdUpdated)
	} else if state == INSERTED {
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.SteamIdInserted)
	}
}

// funkcja wyciągająca za pomocą specjalnego api steamId na podstawie id konta użytkownika. Zwraca również potencjalny błąd.
func validateSteamId(steamId string) error {
	// pierwsze zapytanie robimy do loga tego api, ponieważ tego wymagają w swoich zasadach korzystania z API.
	log.Println("Zapisuję do logu SID " + steamId)
	url := fmt.Sprintf("https://steamid.co/php/log.php?link=%s", steamId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("getSteamIdForProfileId http.DefaultClient.Do(req) " + err.Error())
		return err
	}
	resp.Body.Close()
	// Drugie zapytanie jest już dokładnie do pozyskania steamId
	log.Println("Sprawdzam zgodność SID " + steamId)
	url = fmt.Sprintf("https://steamid.co/php/api.php?action=steamID64&id=%s", steamId)
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Println("getSteamIdForProfileId http.DefaultClient.Do(req2) " + err.Error())
		return err
	}
	defer resp.Body.Close()
	// struktura wewnętrzna, w JSONie który dostaniemy z api może być pole error, lub pole steamId64 (między innymi). Tylko one nas interesują.
	var data struct {
		SteamID64 string `json:"steamID64"`
		Error     string `json:"error"`
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return err
	}
	// Jeżeli error został ustawiony to znaczy że takiego profilu nie ma, upraszczając.
	if data.Error != "" {
		log.Println("Stwierdzam niezgodność SID " + steamId)
		return noSuchProfileError
	}
	log.Println("Stwierdzam zgodność SID " + steamId)
	return nil
}

// funkcja wyciągająca za pomocą specjalnego api steamId na podstawie id konta użytkownika. Zwraca również potencjalny błąd.
func getSteamIdForProfileId(profileId string) (string, error) {
	// pierwsze zapytanie robimy do loga tego api, ponieważ tego wymagają w swoich zasadach korzystania z API.
	log.Println("Zapisuję do logu SID " + profileId)
	url := fmt.Sprintf("https://steamid.co/php/log.php?link=%s", profileId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("getSteamIdForProfileId http.DefaultClient.Do(req) " + err.Error())
		return "", err
	}
	resp.Body.Close()
	// Drugie zapytanie jest już dokładnie do pozyskania steamId
	log.Println("Sprawdzam zgodność " + profileId)
	url = fmt.Sprintf("https://steamid.co/php/api.php?action=steamID&id=%s", profileId)
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Println("getSteamIdForProfileId http.DefaultClient.Do(req2) " + err.Error())
		return "", err
	}
	defer resp.Body.Close()
	// struktura wewnętrzna, w JSONie który dostaniemy z api może być pole error, lub pole steamId64 (między innymi). Tylko one nas interesują.
	var data struct {
		SteamID64 string `json:"steamID64"`
		Error     string `json:"error"`
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", err
	}
	// Jeżeli error został ustawiony to znaczy że takiego profilu nie ma, upraszczając.
	if data.Error != "" {
		log.Println("Stwierdzam niezgodność " + profileId)
		return "", noSuchProfileError
	}
	log.Println("Stwierdzam zgodność " + profileId)
	return data.SteamID64, nil
}

//funkcja powiązuje id użytkownika discorda z steamid w bazie danych
func linkUser(discordID, steamID string) State {
	var linkedUser LinkedUsers
	// sprawdzamy, czy takie id discorda jest już powiązane, unikamy duplikatów ,aktualizujemy.
	err := DbMap.SelectOne(&linkedUser, "SELECT * FROM LinkedUsers WHERE discord_id=?", discordID)
	linkedUser.ExpirationDate = time.Now().Add(24 * time.Hour)
	// jeżeli nie ma wpisu z takim discord id...
	if err == sql.ErrNoRows {
		linkedUser.DiscordID = discordID
		linkedUser.SteamID64 = steamID
		linkedUser.Valid = true
		err = DbMap.Insert(&linkedUser)
		if err != nil {
			log.Println("Błąd połączenia z bazą danych!\n" + err.Error())
			return ERROR
		}
		return INSERTED
	}
	if err != nil {
		log.Println("Błąd połączenia z bazą danych!\n" + err.Error())
		return ERROR
	}
	//a jeżeli takowy wpis jest
	linkedUser.SteamID64 = steamID
	// dla pewności
	linkedUser.Valid = true
	_, err = DbMap.Update(&linkedUser)
	if err != nil {
		log.Println("Błąd połączenia z bazą danych!\n" + err.Error())
		return ERROR
	}
	return UPDATED
}

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

//funkcja odpalana cyklicznie, sprawdza czy wszyscy w bazie nadal są na serwerze i czy są na nim vipami.
func checkUsers() {
	roleId, err := getRoleID(Config.ServerId, Config.PermittedRoleName)
	if err != nil {
		log.Println("Błąd pobierania informacji o roli!\n" + err.Error())
		return
	}
	//zbieramy wszystkie zapisane ID z discorda
	var linkedUsers []LinkedUsers
	_, err = DbMap.Select(&linkedUsers, "SELECT discord_id,valid FROM LinkedUsers")
	var stay bool
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
		stay = false
		for _, role := range member.Roles {
			if role == roleId {
				stay = true
				break
			}
		}
		//jeżeli użytkownik jest na serwerze, ale nie ma już tej rangi - odcinamy go.
		if stay == false {
			if linkedUser.Valid == true {
				_, _ = DbMap.Exec("UPDATE LinkedUsers SET valid=false WHERE discord_id=?", linkedUser.DiscordID)
			}
		} else {
			// jeżeli jednak był
			if linkedUser.Valid == false {
				_, _ = DbMap.Exec("UPDATE LinkedUsers SET valid=true WHERE discord_id=?", linkedUser.DiscordID)
			}
		}
	}
}
