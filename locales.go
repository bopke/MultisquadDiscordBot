package main

import (
	"encoding/json"
	"log"
	"os"
)

//struktura z tekstami
type Locales struct {
	NoPermission                  string `json:"no_permission"`
	NoAdminPermission             string `json:"no_admin_permission"`
	RateLimitWait                 string `json:"rate_limit_wait"`
	SteamInvalidProfileLink       string `json:"steam_invalid_profile_link"`
	SteamInvalidProfileId         string `json:"steam_invalid_profile_id"`
	SteamInstruction              string `json:"steam_instruction"`
	DatabaseError                 string `json:"database_error"`
	SteamIdUpdated                string `json:"steamid_updated"`
	SteamIdInserted               string `json:"steamid_inserted"`
	UnexpectedApiError            string `json:"unexpected_api_error"`
	MinecraftIncorrectNickname    string `json:"minecraft_incorrect_nickname"`
	MinecraftInsertedNickname     string `json:"minecraft_inserted_nickname"`
	MinecraftUpdatedNickname      string `json:"minecraft_updated_nickname"`
	MinecraftInstruction          string `json:"minecraft_instruction"`
	StatusNoVip                   string `json:"status_no_vip"`
	StatusExpired                 string `json:"status_expired"`
	StatusValid                   string `json:"status_valid"`
	VipIncorrectDaysCount         string `json:"vip_incorrect_days_count"`
	VipIncorrectUser              string `json:"vip_incorrect_user"`
	VipInserted                   string `json:"vip_inserted"`
	VipUpdated                    string `json:"vip_updated"`
	VipAnnouncementInserted       string `json:"vip_announcement_inserted"`
	VipAnnouncementUpdated        string `json:"vip_announcement_updated"`
	VipNearExpirationNotification string `json:"vip_near_expiration_notification"`
	VipExpiredNotification        string `json:"vip_expired_notification"`
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
	l.log()
	return
}

func (l *Locales) log() {
	log.Println("===== LOCALES =====")
	log.Println(" No permission: " + Locale.NoPermission)
	log.Println(" Steam invalid profile link: " + Locale.SteamInvalidProfileLink)
	log.Println(" Steam invalid profile id: " + Locale.SteamInvalidProfileId)
	log.Println(" Steam instruction: " + Locale.SteamInstruction)
	log.Println(" Database error: " + Locale.DatabaseError)
	log.Println(" Steam id updated: " + Locale.SteamIdUpdated)
	log.Println(" Steam id inserted: " + Locale.SteamIdInserted)
	log.Println(" Unexpected api error: " + Locale.UnexpectedApiError)
	log.Println(" Minecraft incorrect nickname: " + Locale.MinecraftIncorrectNickname)
	log.Println(" Minecraft inserted nickname: " + Locale.MinecraftInsertedNickname)
	log.Println(" Minecraft updated nickname: " + Locale.MinecraftUpdatedNickname)
	log.Println(" Minecraft instruction: " + Locale.MinecraftInstruction)
	log.Println(" Status no vip: " + Locale.StatusNoVip)
	log.Println(" Status expired: " + Locale.StatusExpired)
	log.Println("Status valid: " + Locale.StatusValid)
	log.Println("===== LOCALES =====")
}
