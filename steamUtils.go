package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

//error używany przy zwracaniu błędu braku profilu w funkcji wydobywania steamid64
var NoSuchProfileError = errors.New("no such profile")

// funkcja wyciągająca za pomocą specjalnego api steamId na podstawie id konta użytkownika. Zwraca również potencjalny błąd.
func validateSteamId(steamId string) error {
	log.Println("Sprawdzam zgodność SID " + steamId)
	url := fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002?key=%s&steamids=%s", Config.SteamApiToken, steamId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("getSteamIdForProfileId http.NewRequest(\"GET\", url, nil) " + err.Error())
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("getSteamIdForProfileId http.DefaultClient.Do(req) " + err.Error())
		return err
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
		return err
	}
	if len(data.Response.Players) == 0 {
		log.Println("Stwierdzam niezzgodność SID " + steamId)
		return NoSuchProfileError
	}
	log.Println("Stwierdzam zgodność SID " + steamId)
	return nil
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
