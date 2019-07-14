package main

import (
	"encoding/json"
	"log"
	"os"
)

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
