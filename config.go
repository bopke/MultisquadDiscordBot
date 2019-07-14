package main

import (
	"encoding/json"
	"log"
	"os"
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
