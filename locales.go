package main

import (
	"encoding/json"
	"log"
	"os"
)

//struktura z tekstami
type Locales struct {
	NoPermission                    string `json:"no_permission"`
	NoAdminPermission               string `json:"no_admin_permission"`
	RateLimitWait                   string `json:"rate_limit_wait"`
	SteamInvalidProfileLink         string `json:"steam_invalid_profile_link"`
	SteamInvalidProfileId           string `json:"steam_invalid_profile_id"`
	SteamInstruction                string `json:"steam_instruction"`
	DatabaseError                   string `json:"database_error"`
	SteamIdUpdated                  string `json:"steamid_updated"`
	SteamIdInserted                 string `json:"steamid_inserted"`
	UnexpectedApiError              string `json:"unexpected_api_error"`
	MinecraftIncorrectNickname      string `json:"minecraft_incorrect_nickname"`
	MinecraftInsertedNickname       string `json:"minecraft_inserted_nickname"`
	MinecraftUpdatedNickname        string `json:"minecraft_updated_nickname"`
	MinecraftInstruction            string `json:"minecraft_instruction"`
	StatusNoVip                     string `json:"status_no_vip"`
	StatusExpired                   string `json:"status_expired"`
	StatusValid                     string `json:"status_valid"`
	VipIncorrectDaysCount           string `json:"vip_incorrect_days_count"`
	VipIncorrectUser                string `json:"vip_incorrect_user"`
	VipInserted                     string `json:"vip_inserted"`
	VipUpdated                      string `json:"vip_updated"`
	VipAnnouncementInserted         string `json:"vip_announcement_inserted"`
	VipAnnouncementUpdated          string `json:"vip_announcement_updated"`
	VipNearExpirationNotification   string `json:"vip_near_expiration_notification"`
	VipExpiredNotification          string `json:"vip_expired_notification"`
	ColorIncorrectDaysCount         string `json:"color_incorrect_days_count"`
	ColorIncorrectColorCode         string `json:"color_incorrect_color_code"`
	ColorIncorrectUser              string `json:"color_incorrect_user"`
	ColorInserted                   string `json:"color_inserted"`
	ColorUpdated                    string `json:"color_updated"`
	ColorAnnouncementInserted       string `json:"color_announcement_inserted"`
	ColorAnnouncementUpdated        string `json:"color_announcement_updated"`
	ColorNearExpirationNotification string `json:"vip_near_expiration_notification"`
	ColorExpiredNotification        string `json:"vip_expired_notification"`
	ReportStage1Message             string `json:"report_stage_1_message"`
	ReportStage2Message             string `json:"report_stage_2_message"`
	ReportStage3Message             string `json:"report_stage_3_message"`
	ReportStage4Message             string `json:"report_stage_4_message"`
	ReportConfirmedMessage          string `json:"report_confirmed_message"`
	ReportDeclinedMessage           string `json:"report_declined_message"`
	ErrorCreatingDMChannel          string `json:"error_creating_dm_channel"`
	ErrorNoReportedUser             string `json:"error_no_reported_user"`
	RaidConfirmation                string `json:"raid_confirmation"`
	RaidConfirmationTimed           string `json:"raid_confirmation_timed"`
	RaidConfirmed                   string `json:"raid_confirmed"`
	RaidConfirmedTimed              string `json:"raid_confirmed_timed"`
	RaidEndedTimed                  string `json:"raid_ended_timed"`
	RaidEnded                       string `json:"raid_ended"`
	RaidRefused                     string `json:"raid_refused"`
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
