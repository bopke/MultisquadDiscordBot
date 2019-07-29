package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"net/http"
	"strings"
)

//error używany przy zwracaniu błędu braku profilu w funkcji wydobywania steamid64
var NoSuchProfileError = errors.New("no such profile")

// funkcja wyciągająca za pomocą specjalnego api steamId na podstawie id konta użytkownika. Zwraca również potencjalny błąd.
func validateSteamId(steamId string) (string, error) {
	log.Println("Sprawdzam zgodność SID " + steamId)
	url := fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002?key=%s&steamids=%s", Config.SteamApiToken, steamId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("getSteamIdForProfileId http.NewRequest(\"GET\", url, nil) " + err.Error())
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("getSteamIdForProfileId http.DefaultClient.Do(req) " + err.Error())
		return "", err
	}
	defer resp.Body.Close()
	// struktura wewnętrzna, w JSONie który dostaniemy z api może być pole error, lub pole steamId64 (między innymi). Tylko one nas interesują.
	type players struct {
		SteamId string `json:"steamid"`
	}
	type response struct {
		Players []players `json:"players"`
	}
	var data struct {
		Response response `json:"response"`
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		log.Println("getSteamIdForProfileId json.Decode " + err.Error())
		return "", err
	}
	if len(data.Response.Players) == 0 {
		log.Println("Stwierdzam niezzgodność SID " + steamId)
		return "", NoSuchProfileError
	}
	log.Println("Stwierdzam zgodność SID " + steamId)
	return data.Response.Players[0].SteamId, nil
}

// funkcja wyciągająca za pomocą specjalnego api steamId na podstawie id konta użytkownika. Zwraca również potencjalny błąd.
func getSteamIdForProfileId(profileId string) (string, error) {
	log.Println("Sprawdzam zgodność " + profileId)
	url := fmt.Sprintf("http://api.steampowered.com/ISteamUser/ResolveVanityURL/v0001?key=%s&vanityurl=%s", Config.SteamApiToken, profileId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("getSteamIdForProfileId http.DefaultClient.Do(req) " + err.Error())
		return "", err
	}
	defer resp.Body.Close()
	// struktura wewnętrzna JSONa
	type response struct {
		SteamId string `json:"steamid"`
	}
	var data struct {
		Response response `json:"response"`
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", err
	}
	if data.Response.SteamId == "" {
		log.Println("Stwierdzam niezgodność " + profileId)
		return "", NoSuchProfileError
	}
	log.Println("Stwierdzam zgodność " + profileId)
	return data.Response.SteamId, nil
}

//funkcja powiązuje id użytkownika discorda z steamid w bazie danych
func linkUserSteamID(discordID, steamID string) State {
	var linkedUser LinkedUsers
	// sprawdzamy, czy takie id discorda jest już powiązane, unikamy duplikatów ,aktualizujemy.
	err := DbMap.SelectOne(&linkedUser, "SELECT * FROM LinkedUsers WHERE discord_id=?", discordID)
	// jeżeli nie ma wpisu z takim discord id...
	if err == sql.ErrNoRows {
		linkedUser.DiscordID = discordID
		linkedUser.SteamID64.String = steamID
		linkedUser.SteamID64.Valid = true
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
	linkedUser.SteamID64.String = steamID
	linkedUser.SteamID64.Valid = true
	// dla pewności
	linkedUser.Valid = true
	_, err = DbMap.Update(&linkedUser)
	if err != nil {
		log.Println("Błąd połączenia z bazą danych!\n" + err.Error())
		return ERROR
	}
	return UPDATED
}

func handleSteamCommand(s *discordgo.Session, message *discordgo.MessageCreate) {
	member, err := s.GuildMember(message.GuildID, message.Author.ID)
	if err != nil {
		log.Println("Błąd pobierania twórcy wiadomości!\n" + err.Error())
		return
	}
	if !hasRole(member, Config.PermittedRoleName, message.GuildID) {
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.NoPermission)
		return
	}
	// dzielimy wiadomość po spacjach dla wygody
	args := strings.Split(message.Content, " ")
	if len(args) < 2 {
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.SteamInstruction)
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
			args[1], err = validateSteamId(args[1])
			if err == nil {
				state := linkUserSteamID(message.Author.ID, args[1])
				if state == ERROR {
					_, _ = s.ChannelMessageSend(message.ChannelID, Locale.DatabaseError)
				} else if state == UPDATED {
					_, _ = s.ChannelMessageSend(message.ChannelID, Locale.SteamIdUpdated)
				} else if state == INSERTED {
					_, _ = s.ChannelMessageSend(message.ChannelID, Locale.SteamIdInserted)
				}
				return
			} else if err != NoSuchProfileError {
				_, _ = s.ChannelMessageSend(message.ChannelID, Locale.UnexpectedApiError)
				return
			}
		}
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.SteamInvalidProfileLink)
		return
	}
	//interesuje nas tylko to, co jest po id/
	args[1] = args[1][22:]
	// i na podstawie tego możemy pobrać steamID
	steamId, err := getSteamIdForProfileId(args[1])
	if err != nil {
		// ale ktoś może podać zły link i taki profil nie istnieje!
		if err == NoSuchProfileError {
			_, _ = s.ChannelMessageSend(message.ChannelID, Locale.SteamInvalidProfileId)
			return
		}
		// może też wystąpić jakiś nieoczekiwany błąd..
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.UnexpectedApiError)
		log.Println("Błąd przy odczycie danych z API!\n" + err.Error())
		return
	}
	state := linkUserSteamID(message.Author.ID, steamId)
	if state == ERROR {
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.DatabaseError)
	} else if state == UPDATED {
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.SteamIdUpdated)
	} else if state == INSERTED {
		_, _ = s.ChannelMessageSend(message.ChannelID, Locale.SteamIdInserted)
	}

}
